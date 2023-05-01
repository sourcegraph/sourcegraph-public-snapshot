import numpy as np

np.random.seed(0)


def print_as_array(name, arr):
    print("var", name, "= []float32{")
    for row in arr:
        print("\t", end="")
        for col in row:
            print(f"{col:.4f}, ", end="")
        print()
    print("}")


def print_as_2d_array(name, arr):
    print("var", name, "= [][]int{")
    for row in arr:
        print("\t{", end="")
        for col in row:
            print(f"{col}, ", end="")
        print("},")
    print("}")


N_ROWS, N_COLS, N_QUERIES = 16, 3, 3

A = np.random.random((N_ROWS, N_COLS))
A_norm = A / np.linalg.norm(A, axis=1, keepdims=True)

q = np.random.random((N_QUERIES, N_COLS))
q_norm = q / np.linalg.norm(q, axis=1, keepdims=True)

print_as_array("embeddings", A_norm)
print_as_array("queries", q_norm)

similarities = np.matmul(q_norm, A_norm.T)
similarity_ranks = (N_ROWS - 1) - np.argsort(
    np.argsort(similarities, axis=1),
    axis=1,
)
print_as_2d_array("ranks", similarity_ranks)
