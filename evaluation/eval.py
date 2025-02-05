import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import argparse

# Set up argument parser
parser = argparse.ArgumentParser(description='Evaluate benchmark results and plot latency over time.')
parser.add_argument('--output', type=str, required=True, help='Output plot filename')
args = parser.parse_args()

# Load CSV
df = pd.read_csv("benchmark_durable50.csv", header=None)

# Assign correct column names
df.columns = ["message_id", "sent_time", "received_time", "benchmark_id", "queue_type"]

# Convert timestamps to datetime (handling RFC3339Nano format)
df['sent_time'] = pd.to_datetime(df['sent_time'], format="%Y-%m-%dT%H:%M:%S.%f%z", utc=True)
df['received_time'] = pd.to_datetime(df['received_time'], format="%Y-%m-%dT%H:%M:%S.%f%z", utc=True)

# Compute latency (in milliseconds)
df['latency_ms'] = (df['received_time'] - df['sent_time']).dt.total_seconds() * 1000

# Compute average latency per second
df['sent_time_rounded'] = df['sent_time'].dt.floor('s')  # Round to nearest second
df_mean = df.groupby('sent_time_rounded')['latency_ms'].mean().reset_index()

# Group by second and count messages
df["sent_time_rounded"] = df["sent_time"].dt.floor('s')  # Round to nearest second
throughput = df.groupby("sent_time_rounded")["message_id"].count().reset_index()

# Rename column for clarity
throughput.columns = ["time", "messages_per_second"]

plt.figure(figsize=(10, 5))
sns.lineplot(x=df_mean['sent_time_rounded'], y=df_mean['latency_ms'])
plt.xlabel("Time (per second)")
plt.ylabel("Mean Latency (ms)")
plt.title("Mean Message Latency Over Time")
plt.xticks(rotation=45)
plt.grid(True)

# Save plot
plt.savefig("./plots/" + args.output + "_latency.png")

# Plot throughput over time
plt.figure(figsize=(10, 5))
plt.plot(throughput["time"], throughput["messages_per_second"], marker='o', linestyle='-')
plt.xlabel("Time (s)")
plt.ylabel("Throughput (Messages per Second)")
plt.title("Message Throughput Over Time")
plt.xticks(rotation=45)
plt.grid(True)

# Save plot
plt.savefig("./plots/" + args.output + "_throughput.png")
