import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer



import sys
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("command", type=str)
parser.add_argument("--clusters", type=int)
parser.add_argument("--tsv_text", type=str)
args = parser.parse_args()

df = pd.read_csv(args.tsv_text, sep='\t')

# Initialize a TF-IDF vectorizer
tfidf_vectorizer = TfidfVectorizer(max_df=0.85, max_features=1000, stop_words='english')

# Fit and transform the complaints text to create embeddings
tfidf_matrix = tfidf_vectorizer.fit_transform(df['text'])

#tfidf_matrix.shape

from sklearn.decomposition import PCA

# Initialize PCA and reduce dimensionality to 2 components
pca = PCA(n_components=2)
reduced_tfidf = pca.fit_transform(tfidf_matrix.toarray())

#reduced_tfidf.shape

from sklearn.cluster import KMeans
import matplotlib.pyplot as plt

command = args.command

if command == "elbow":
    # Determine the optimal number of clusters using the Elbow method
    wcss = []  # within-cluster sum of squares
    cluster_range = range(1, 10)  # test up to 10 clusters

    for k in cluster_range:
        kmeans = KMeans(n_clusters=k, random_state=42)
        kmeans.fit(reduced_tfidf)
        wcss.append(kmeans.inertia_)

    # Plot the Elbow method
    plt.figure(figsize=(10, 6))
    plt.plot(cluster_range, wcss, marker='o', linestyle='--')
    plt.xlabel('Number of Clusters')
    plt.ylabel('Within-Cluster Sum of Squares')
    plt.title('Elbow Method for Optimal Number of Clusters')
    plt.grid(True)
    plt.show()
    sys.exit(0)


if command == "cluster":
    num_clusters = args.clusters
    # Perform KMeans clustering with 4 clusters
    kmeans = KMeans(n_clusters=4, random_state=42)
    clusters = kmeans.fit_predict(reduced_tfidf)

    # Plot the clusters
    plt.figure(figsize=(10, 6))
    plt.scatter(reduced_tfidf[:, 0], reduced_tfidf[:, 1], c=clusters, cmap='rainbow')
    plt.scatter(kmeans.cluster_centers_[:, 0], kmeans.cluster_centers_[:, 1], s=200, c='black', marker='X', label='Centroids')
    plt.xlabel('PCA Component 1')
    plt.ylabel('PCA Component 2')
    plt.title('Clusters of English statements')
    plt.legend()
    plt.grid(True)
    plt.show()

    for cluster in range(0, 4):
        print('\ncluster:', cluster)
        for index, row in df.iterrows():
            if clusters[index] == cluster:
                print(row['text'].strip())
