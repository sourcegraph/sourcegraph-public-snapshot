import pandas as pd
from sentence_transformers import SentenceTransformer
from sklearn.decomposition import PCA


import sys
import argparse
import os
os.environ["TOKENIZERS_PARALLELISM"] = "false" #avoids parallelism warnings

parser = argparse.ArgumentParser()
parser.add_argument("--clusters", default=4, type=int, help="Number of clusters to generate")
parser.add_argument("--input", type=str, help="Path to input TSV file containing text")
parser.add_argument("--text_field", default="text", type=str, help="Name of column in TSV containing the text to create embeddings and cluster")
parser.add_argument("--output", type=str, help="Path to output file")
parser.add_argument("--model", default="all-MiniLM-L6-v2", type=str, help="Sentence transformer model name")
parser.add_argument("--quiet",action="store_true", help="hides progress bars and skip displaying plots")

args = parser.parse_args()

# Initialize an embeddings model
embedding_model = SentenceTransformer(args.model)

# Read the input tsv file into a dataframe
df = pd.read_csv(args.input, sep='\t')

# Generate embeddings
embeddings = embedding_model.encode(df[args.text_field], show_progress_bar=not args.quiet,normalize_embeddings=True)

from sklearn.cluster import KMeans
import matplotlib.pyplot as plt


# Perform KMeans clustering
cluster_model = KMeans(n_clusters=args.clusters, random_state=42, n_init='auto')
clusters = cluster_model.fit_predict(embeddings)

# update the dataframe with the cluster assignments
df['cluster'] = cluster_model.labels_


if not args.quiet:
    # plot the clusters
    # Initialize PCA and reduce dimensionality to 2 components
    pca = PCA(n_components=2)
    reduced_embeddings = pca.fit_transform(embeddings)
    plt.figure(figsize=(10, 6))
    plt.scatter(reduced_embeddings[:, 0], reduced_embeddings[:, 1], c=clusters, cmap='rainbow')
    plt.xlabel('PCA Component 1')
    plt.ylabel('PCA Component 2')
    plt.title('Clusters of English statements')
    plt.grid(True)
    plt.show()


# if an output was specified write the dataframe to a tsv file
if args.output is not None:
    df.sort_values(by=["cluster"]).to_csv(args.output, sep='\t', index=False)
