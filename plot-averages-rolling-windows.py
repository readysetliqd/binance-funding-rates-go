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
            FROM top100_historical_funding_rates GROUP BY funding_time, symbol, mark_price;
        '''
        
# Execute the query
cur.execute(query)

# Fetch all rows from the cursor
rows = cur.fetchall()

# Get the column names from the cursor description
columns = [desc[0] for desc in cur.description]

# Create a pandas DataFrame from the fetched data
df = pd.DataFrame(rows, columns=columns)

# Sort the DataFrame
df = df.sort_values(['symbol', 'funding_time'])

# Calculate the percentage change in `mark_price` for each coin
df['returns'] = df.groupby('symbol')['mark_price'].pct_change(fill_method=None)

# Drop rows with NaN values
df.dropna(subset=['returns'], inplace=True)

# Calculate the average return for each `funding_time`
returns_df = df.groupby('funding_time')['returns'].mean().reset_index()

# PostgreSQL query to fetch data from your table
query = '''SELECT funding_time,
            AVG(funding_rate) AS avg_funding_rate, 
            MIN(funding_rate) AS min_funding_rate, 
            PERCENTILE_CONT(0.25) WITHIN GROUP(ORDER BY funding_rate) AS q1_funding_rate, 
            PERCENTILE_CONT(0.5) WITHIN GROUP(ORDER BY funding_rate) AS q2_funding_rate, 
            PERCENTILE_CONT(0.75) WITHIN GROUP(ORDER BY funding_rate) AS q3_funding_rate, 
            MAX(funding_rate) AS max_funding_rate 
            FROM top100_historical_funding_rates GROUP BY funding_time;
        '''
# Execute the query
cur.execute(query)

# Fetch all rows from the cursor
rows = cur.fetchall()

# Get the column names from the cursor description
columns = [desc[0] for desc in cur.description]

# Create a pandas DataFrame from the fetched data
df2 = pd.DataFrame(rows, columns=columns)

# Close the cursor and connection
cur.close()
conn.close()

# Define the size of the rolling window in days
window_sizes = [7, 14, 30, 90, 365]  # 7-day, 30-day, 90-day, and 1-year windows

# Convert window sizes from days to number of data points
window_sizes = [int(window_size * 24 / 8) for window_size in window_sizes]  # Assuming 8-hour intervals

# Initialize an empty DataFrame to store the results
results = pd.DataFrame()

# Loop over the different window sizes
for window_size in window_sizes:
    # Loop over the DataFrame using a rolling window concat to results df
    for i in tqdm(range(window_size, len(df2)), desc='Processing'):
        # Define the training data as the current window
        train = df2.iloc[i-window_size:i].copy()  
        
        # Define the test data as the next observation
        test = df2.iloc[i]
        
        train['avg_funding_rate'] = train['avg_funding_rate'].astype(float)
        
        # Calculate the statistics on the training data
        avg_funding_rate = train['avg_funding_rate'].mean()
        min_funding_rate = train['avg_funding_rate'].min()
        q1_funding_rate = train['avg_funding_rate'].quantile(0.25)
        q2_funding_rate = train['avg_funding_rate'].median()
        q3_funding_rate = train['avg_funding_rate'].quantile(0.75)
        max_funding_rate = train['avg_funding_rate'].max()
        
        # Store the results
        new_row = {
        'window_size': window_size,
        'funding_time': test['funding_time'],
        'avg_funding_rate': avg_funding_rate,
        'min_funding_rate': min_funding_rate,
        'q1_funding_rate': q1_funding_rate,
        'q2_funding_rate': q2_funding_rate,
        'q3_funding_rate': q3_funding_rate,
        'max_funding_rate': max_funding_rate
        }

        # Using concat
        results = pd.concat([results, pd.DataFrame([new_row])], ignore_index=True)

# Merge the returns_df DataFrame with the results DataFrame
results = pd.merge(results, returns_df, on='funding_time', how='inner')

# List of periods for which to calculate forward returns
periods = [1, 3, 9, 21, 42, 84]

# Loop over the periods
for period in periods:
    # Calculate the compounded return
    results[f'forward_{period}_period_return'] = ((1 + results['returns']).rolling(period).apply(np.prod, raw=True) - 1).shift(-period)
    results[f'forward_{period}_total_return'] = results[f'forward_{period}_period_return'] - results['avg_funding_rate'].rolling(period).apply(np.sum, raw=True).shift(-period)

# Drop rows with NaN values
results.dropna(subset=[f'forward_{period}_period_return' for period in periods], inplace=True)

# Convert the 'avg_funding_rate' data to float
results['avg_funding_rate'] = results['avg_funding_rate'].astype(float)

# Calculate the Median Absolute Deviation (MAD) score for each observation
median = results['avg_funding_rate'].median()
mad = np.median(np.abs(results['avg_funding_rate'] - median))
results['mad_score'] = np.abs(results['avg_funding_rate'] - median) / mad

# Define "extreme" values as those above the 95th percentile or below the 5th percentile
low_threshold = results['avg_funding_rate'].quantile(0.02)
high_threshold = results['avg_funding_rate'].quantile(0.98)

# Identify extreme values
results['avg_quantile_high'] = (results['avg_funding_rate'] > high_threshold)
results['avg_quantile_low'] = (results['avg_funding_rate'] < low_threshold)
results['mad_extreme_high'] = (results['mad_score'] > 3) & (results['avg_funding_rate'] > results['q2_funding_rate'])
results['mad_extreme_low'] = (results['mad_score'] > 3) & (results['avg_funding_rate'] < results['q2_funding_rate'])

# Create subplots
fig, axs = plt.subplots(3)

# Plot a histogram of the averages
axs[0].hist(results['avg_funding_rate'], bins=100, edgecolor='black')
axs[0].set_title('Histogram of Averages')
axs[0].set_xlabel('Average Funding Rate')
axs[0].set_ylabel('Frequency')

# Plot a boxplot of the averages
axs[1].boxplot(results['avg_funding_rate'])
axs[1].set_title('Boxplot of Averages')
axs[1].set_ylabel('Average Funding Rate')

axs[2].hist(results['mad_score'], bins=100, edgecolor='black')
axs[2].set_title('Histogram of MAD Scores')
axs[2].set_xlabel('MAD Score')
axs[2].set_ylabel('Frequency')

# Display the plots
plt.tight_layout()
plt.show()
    
# List of extreme conditions
extreme_conditions = ['avg_quantile_high', 'avg_quantile_low', 'mad_extreme_high', 'mad_extreme_low']
extreme_conditions_pairs = [('avg_quantile_high', 'avg_quantile_low'), ('mad_extreme_high', 'mad_extreme_low')]

# Loop over the periods to perform t-statistic tests
with open('stats_output.txt', 'w') as f:
    f.write('Price returns ex funding rates\n******************************\n')
    for period in periods:
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Get the forward returns for each group
            condition_true = results[results[condition] == True][f'forward_{period}_period_return']
            condition_false = results[results[condition] == False][f'forward_{period}_period_return']
            
            # Perform a t-test
            t_stat, p_value = stats.ttest_ind(condition_true, condition_false, nan_policy='omit')
            
            f.write(f'T-test for forward {period} period return when {condition} is True:\n')
            f.write(f't-statistic: {t_stat}\n')
            f.write(f'p-value: {p_value}\n')

    # Loop over the periods to perform correlation tests
    for period in periods:
        f.write(f'Correlation with forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Calculate the correlation
            correlation = results[condition].corr(results[f'forward_{period}_period_return'])
            f.write(f'{condition}: {correlation}\n')

    # Loop over the periods to perform regression analysis
    for period in periods:
        f.write(f'Regression analysis for forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
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

    # Loop over the periods to perform confidence intervals tests
    for period in periods:
        f.write(f'Confidence intervals for forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Get the forward returns for the condition
            returns = results[results[condition] == True][f'forward_{period}_period_return']
            # Calculate the mean and standard error
            mean = returns.mean()
            se = stats.sem(returns)
            # Calculate the confidence interval
            ci = stats.t.interval(confidence_level, len(returns)-1, loc=mean, scale=se)
            f.write(f'{condition}: {ci}\n')
            
    # Loop over the periods
    for period in periods:
        f.write(f'Effect size for forward {period} period return:\n')
        # Loop over the pairs of extreme conditions
        for condition_true, condition_false in extreme_conditions_pairs:
            # Get the forward returns for each condition
            group_true = results[results[condition_true] == True][f'forward_{period}_period_return']
            group_false = results[results[condition_false] == True][f'forward_{period}_period_return']
            
            # Calculate the means
            mean_true = group_true.mean()
            mean_false = group_false.mean()
            
            # Calculate the standard deviation of the combined groups
            std_combined = np.sqrt((group_true.var() + group_false.var()) / 2)
            
            # Calculate Cohen's d
            cohens_d = (mean_true - mean_false) / std_combined
            
            f.write(f"Cohen's d for {condition_true} vs {condition_false}: {cohens_d}\n")
        
    f.write('Total returns including funding rates\n******************************\n')
    for period in periods:
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Get the forward returns for each group
            condition_true = results[results[condition] == True][f'forward_{period}_total_return']
            condition_false = results[results[condition] == False][f'forward_{period}_total_return']
            
            # Perform a t-test
            t_stat, p_value = stats.ttest_ind(condition_true, condition_false, nan_policy='omit')
            
            f.write(f'T-test for forward {period} period return when {condition} is True:\n')
            f.write(f't-statistic: {t_stat}\n')
            f.write(f'p-value: {p_value}\n')

    # Loop over the periods to perform correlation tests
    for period in periods:
        f.write(f'Correlation with forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Calculate the correlation
            correlation = results[condition].corr(results[f'forward_{period}_total_return'])
            f.write(f'{condition}: {correlation}\n')

    # Loop over the periods to perform regression analysis
    for period in periods:
        f.write(f'Regression analysis for forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Define the dependent variable (forward return)
            Y = results[f'forward_{period}_total_return']
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

    # Loop over the periods to perform confidence intervals tests
    for period in periods:
        f.write(f'Confidence intervals for forward {period} period return:\n')
        # Loop over the extreme conditions
        for condition in extreme_conditions:
            # Get the forward returns for the condition
            returns = results[results[condition] == True][f'forward_{period}_total_return']
            # Calculate the mean and standard error
            mean = returns.mean()
            se = stats.sem(returns)
            # Calculate the confidence interval
            ci = stats.t.interval(confidence_level, len(returns)-1, loc=mean, scale=se)
            f.write(f'{condition}: {ci}\n')
            
    # Loop over the periods
    for period in periods:
        f.write(f'Effect size for forward {period} period return:\n')
        # Loop over the pairs of extreme conditions
        for condition_true, condition_false in extreme_conditions_pairs:
            # Get the forward returns for each condition
            group_true = results[results[condition_true] == True][f'forward_{period}_total_return']
            group_false = results[results[condition_false] == True][f'forward_{period}_total_return']
            
            # Calculate the means
            mean_true = group_true.mean()
            mean_false = group_false.mean()
            
            # Calculate the standard deviation of the combined groups
            std_combined = np.sqrt((group_true.var() + group_false.var()) / 2)
            
            # Calculate Cohen's d
            cohens_d = (mean_true - mean_false) / std_combined
            
            f.write(f"Cohen's d for {condition_true} vs {condition_false}: {cohens_d}\n")

# Write DataFrame to a CSV file
results.to_csv('aggregated_funding_df.csv', index=False)