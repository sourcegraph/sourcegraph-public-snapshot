import pandas as pd
from sentence_transformers import SentenceTransformer
from sklearn.decomposition import PCA


import sys
import argparse
import os
os.environ["TOKENIZERS_PARALLELISM"] = "false" #avoids parallelism warnings

parser = argparse.ArgumentParser()
parser.add_argument("--clusters", default=4, type=int)
parser.add_argument("--tsv_text", type=str)
parser.add_argument("--text_field", default="text", type=str)
parser.add_argument("--output", type=str)
parser.add_argument("--model", default="all-MiniLM-L6-v2", type=str)

args = parser.parse_args()
if args.clusters is None:
    args.clusters = 4 # default value

if args.text_field is None:
    args.text_field = "text" # default value

if args.model is None:
    args.model = "all-MiniLM-L6-v2" # default value


num_clusters = args.clusters
output_file = args.output
input_file = args.tsv_text
text_field = args.text_field
model = args.model

# Initialize an embeddings model
embedding_model = SentenceTransformer(model)

# Read the input tsv file into a dataframe
df = pd.read_csv(input_file, sep='\t')

# Generate embeddings
embeddings = embedding_model.encode(df[text_field], show_progress_bar=True,normalize_embeddings=True)

from sklearn.cluster import KMeans
import matplotlib.pyplot as plt


# Perform KMeans clustering
cluster_model = KMeans(n_clusters=num_clusters, random_state=42, n_init='auto')
clusters = cluster_model.fit_predict(embeddings)

# update the dataframe with the cluster assignments
df['cluster'] = cluster_model.labels_

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
