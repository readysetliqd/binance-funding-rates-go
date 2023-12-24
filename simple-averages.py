import os

from dotenv import load_dotenv
import matplotlib.pyplot as plt
import numpy as np
import psycopg2
import pandas as pd
import scipy.stats as stats
import statsmodels.api as sm
from tqdm import tqdm



# Load environment variables from .env file
load_dotenv('db.env')

# Establish a connection to the PostgreSQL database
conn = psycopg2.connect(
    dbname=os.getenv('DB_NAME'),
    user=os.getenv('DB_USER'),
    password=os.getenv('DB_PASS'),
    host=os.getenv('DB_HOST'),
    port=os.getenv('DB_PORT')
)


# Create a cursor object
cur = conn.cursor()

# PostgreSQL query to fetch data from your table
query = '''SELECT funding_time, symbol, mark_price
            FROM top10_historical_funding_rates GROUP BY funding_time, symbol, mark_price;
        '''
        
# Execute the query
cur.execute(query)

# Fetch all rows from the cursor
rows = cur.fetchall()

# Get the column names from the cursor description
columns = [desc[0] for desc in cur.description]

# Create a pandas DataFrame from the fetched data
df = pd.DataFrame(rows, columns=columns)
df = df.sort_values(['symbol', 'funding_time'])

# Calculate the percentage change in `mark_price` for each coin
df['avg_returns'] = df.groupby('symbol')['mark_price'].pct_change(fill_method=None)

# Drop rows with NaN values
df.dropna(subset=['avg_returns'], inplace=True)

# Calculate the average return for each `funding_time`
returns_df = df.groupby('funding_time')['avg_returns'].mean().reset_index()
returns_df = returns_df.sort_values('funding_time')

# PostgreSQL query to fetch data from your table
query = '''SELECT funding_time,
            AVG(funding_rate) AS avg_funding_rate,
            PERCENTILE_CONT(0.5) WITHIN GROUP(ORDER BY funding_rate) AS med_funding_rate
            FROM top10_historical_funding_rates GROUP BY funding_time;
        '''
        
# Execute the query
cur.execute(query)

# Fetch all rows from the cursor
rows = cur.fetchall()

# Get the column names from the cursor description
columns = [desc[0] for desc in cur.description]

# Create a pandas DataFrame from the fetched data
df2 = pd.DataFrame(rows, columns=columns)
df2 = df2.sort_values('funding_time')

# Close the cursor and connection
cur.close()
conn.close()

# Initialize an empty DataFrame to store the results
results = pd.merge(df2, returns_df, on='funding_time', how='inner')
results = results.sort_values('funding_time')

# List of periods for which to calculate forward returns
periods = [1, 3, 9, 21, 42, 84]

for period in periods:
    # Calculate the compounded return
    results[f'forward_{period}_period_return'] = ((1 + results['avg_returns']).rolling(period).apply(np.prod, raw=True) - 1).shift(-period)
    results[f'forward_{period}_total_return'] = results[f'forward_{period}_period_return'] - results['avg_funding_rate'].rolling(period).apply(np.sum, raw=True).shift(-period)

# Drop rows with NaN values
results.dropna(subset=[f'forward_{period}_period_return' for period in periods], inplace=True)

# Convert the 'avg_funding_rate' data to float
results['avg_funding_rate'] = results['avg_funding_rate'].astype(float)
results['med_funding_rate'] = results['med_funding_rate'].astype(float)

results['extreme_high_0.05'] = (results['med_funding_rate'] > 0.0005)
results['extreme_high_0.10'] = (results['med_funding_rate'] > 0.0010)
results['extreme_low_0.05'] = (results['med_funding_rate'] < -0.0005)
results['extreme_low_0.10'] = (results['med_funding_rate'] < -0.0010)

extreme_conditions = ['extreme_high_0.05','extreme_high_0.10','extreme_low_0.05','extreme_low_0.10']
extreme_conditions_pairs = [('extreme_high_0.05','extreme_high_0.10'),('extreme_low_0.05','extreme_low_0.10')]

with open('simple_stats_output.txt', 'w') as f:
    f.write('******************************\nPrice Returns ex Funding Rates\n******************************\n')
    for period in periods:
        f.write(f"\n\n==============================================================================\nForward {period} Period Returns\n==============================================================================")
        # Loop over the extreme conditions
        for condition_pair in extreme_conditions_pairs:
            for condition in condition_pair:
                f.write(f"\n----------------------------------------------------\nForward {period} Period Returns Statistics When {condition} is True\n----------------------------------------------------\n")
                # Get the forward returns for each group
                condition_true = results[results[condition] == True][f'forward_{period}_period_return']
                condition_false = results[results[condition] == False][f'forward_{period}_period_return']

                # Perform a t-test
                t_stat, p_value = stats.ttest_ind(condition_true, condition_false, nan_policy='omit')

                f.write(f'***********T-test***********\n')
                f.write(f't-statistic: {t_stat}\n')
                f.write(f'p-value: {p_value}\n')
                f.write("\n")
                
                # Perform correlation tests
                f.write(f'***********Correlation***********\n')

                # Calculate the correlation
                correlation = results[condition].corr(results[f'forward_{period}_period_return'])
                f.write(f'{condition}: {correlation}\n')
                f.write("\n")

                # Perform regression analysis
                f.write(f'*********** OLS Regression Analysis***********\n')
                # Define the dependent variable (forward return)
                Y = results[f'forward_{period}_period_return']
                # Define the independent variable (extreme condition)
                X = results[condition]
                X = X.astype(int)
                # Add a constant to the independent variable
                X = sm.add_constant(X)
                # Perform the regression analysis
                model = sm.OLS(Y, X, missing='drop')
                fit_results = model.fit()
                f.write(fit_results.summary().as_text())
                f.write("\n")
            
                # Confidence level
                confidence_level = 0.95
                # Perform confidence intervals tests
                f.write(f'\n***********Confidence Intervals***********\n')
                # Get the forward returns for the condition
                returns = results[results[condition] == True][f'forward_{period}_period_return']
                # Calculate the mean and standard error
                mean = returns.mean()
                se = stats.sem(returns)
                # Calculate the confidence interval
                ci = stats.t.interval(confidence_level, len(returns)-1, loc=mean, scale=se)
                f.write(f'{condition}: {ci}\n')
                f.write("\n")
                
                # Function to calculate Cohen's d
                def cohens_d(group1, group2):
                    # Calculate the size of samples
                    n1, n2 = len(group1), len(group2)
                    # Calculate the variance of the samples
                    s1, s2 = np.var(group1, ddof=1), np.var(group2, ddof=1)
                    # Calculate the pooled standard deviation
                    s = np.sqrt(((n1 - 1) * s1 + (n2 - 1) * s2) / (n1 + n2 - 2))
                    # Calculate Cohen's d
                    u1, u2 = np.mean(group1), np.mean(group2)
                    return (u1 - u2) / s
                extreme_returns = results[results[condition]][f'forward_{period}_period_return']
                no_extreme_returns = results[~results[condition]][f'forward_{period}_period_return']
                # Calculate Cohen's d for extreme vs no extreme
                d = cohens_d(extreme_returns, no_extreme_returns)
                f.write(f"***********Cohen's D ***********\n")
                f.write(f"{condition}: {d}\n")