import os

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import pymongo

EXPERIMENT_NAME = "baseline_1k"
ROUTER_REPLICAS = "2"

# Connect to MongoDB
client = pymongo.MongoClient(os.getenv("MONGO_URI"))
db = client["tcc-telemetry"]
collection = db["metrics"]


def make_histogram(experiment_name: str, router_replicas: str):
    query = {
        "tags.experiment": experiment_name,
        "tags.router_replicas": router_replicas,
        # "tags.router_replicas": {"$in": ["1", "2"]},
    }

    projection = {"value": 1, "tags": 1, "_id": 0}

    # Execute query and convert to DataFrame
    cursor = collection.find(query, projection)
    df = pd.DataFrame(list(cursor))
    # print(df.head())

    df["router_replicas"] = df["tags"].apply(lambda x: x.get("router_replicas"))

    df = df.drop(columns=["tags"])

    # Histogram
    plt.figure(figsize=(12, 7))

    # Calculate bins based on data range
    all_values = df["value"]
    bin_min = max(0, all_values.min() - 5)
    bin_max = all_values.max() + 5
    bins = np.linspace(bin_min, bin_max, 50)

    # Plot histograms
    plt.hist(
        df[df["router_replicas"] == router_replicas]["value"],
        bins=bins,
        alpha=0.7,
        # label="1 Router Replica",
        color="blue",
        edgecolor="black",
    )

    # Configure plot
    plt.title(
        f"Execution DeltaTime Distribution\nExperiment: {experiment_name}\nRouterReplicas: {router_replicas}",
        fontsize=14,
    )
    plt.xlabel("DeltaTime (ms)", fontsize=12)
    plt.ylabel("Frequency", fontsize=12)
    plt.legend(fontsize=11)
    plt.grid(axis="y", alpha=0.3)
    plt.gca().set_axisbelow(True)  # Grid behind bars

    plt.tight_layout()
    plt.savefig(
        f"histogran_{experiment_name}_{router_replicas}_router_replicas.png", dpi=300
    )
    plt.show()


def main():
    make_histogram(EXPERIMENT_NAME, ROUTER_REPLICAS)


if __name__ == "__main__":
    main()
