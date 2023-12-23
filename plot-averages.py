import psycopg2
import pandas as pd
import matplotlib.pyplot as plt
from dotenv import load_dotenv
import os
import numpy as np
from scipy.stats import mannwhitneyu


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
df = pd.DataFrame(rows, columns=columns)

# Close the cursor and connection
cur.close()
conn.close()

# Convert the 'avg_funding_rate' data to float
df['avg_funding_rate'] = df['avg_funding_rate'].astype(float)

# Calculate the Median Absolute Deviation (MAD) score for each observation
median = df['avg_funding_rate'].median()
mad = np.median(np.abs(df['avg_funding_rate'] - median))
df['mad_score'] = np.abs(df['avg_funding_rate'] - median) / mad

# Define "extreme" values as those above the 95th percentile or below the 5th percentile
low_threshold = df['avg_funding_rate'].quantile(0.05)
high_threshold = df['avg_funding_rate'].quantile(0.95)
low_mad_threshold = df['mad_score'].quantile(0.05)
high_mad_threshold = df['mad_score'].quantile(0.95)

# Identify extreme values
df['extreme_avg'] = (df['avg_funding_rate'] < low_threshold) | (df['avg_funding_rate'] > high_threshold)
df['extreme_mad'] = (df['mad_score'] < low_mad_threshold) | (df['mad_score'] > high_mad_threshold)

# Create subplots
fig, axs = plt.subplots(3)

# Plot a histogram of the averages
axs[0].hist(df['avg_funding_rate'], bins=30, edgecolor='black')
axs[0].set_title('Histogram of Averages')
axs[0].set_xlabel('Average Funding Rate')
axs[0].set_ylabel('Frequency')

# Plot a boxplot of the averages
axs[1].boxplot(df['avg_funding_rate'])
axs[1].set_title('Boxplot of Averages')
axs[1].set_ylabel('Average Funding Rate')

axs[2].hist(df['mad_score'], bins=30, edgecolor='black')
axs[2].set_title('Histogram of MAD Scores')
axs[2].set_xlabel('MAD Score')
axs[2].set_ylabel('Frequency')
# Display the plots
plt.tight_layout()
plt.show()

# Print all rows of the DataFrame to check the identified extreme values
print(df[(df['extreme_avg'] == True) | (df['extreme_mad'] == True)])

# Write DataFrame to a CSV file
df.to_csv('stats_output.csv', index=False)