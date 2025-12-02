# Smart Copy

Currently, the copy operation removes the entire destination directory and copies everything from scratch. This is inefficient for large directories.

## Approaches

### Approach 1: ModTime + Size comparison

Compare `ModTime` and `Size` of source vs destination files:
- If destination doesn't exist → copy
- If ModTime or Size differ → copy
- Otherwise → skip

**Pros:**
- Simple to implement
- Very fast (no file content reading)
- Same approach used by `make`, `rsync --update`

**Cons:**
- Won't detect changes if someone manipulates timestamps manually (rare edge case)

### Approach 2: Hybrid (ModTime + Size + Hash)

First check ModTime and Size (fast path), then verify with content hash if they differ:
- If ModTime and Size are equal → skip
- If they differ → compute hash (SHA256 or xxHash) to confirm actual content change

**Pros:**
- Balance between speed and accuracy
- Avoids unnecessary copies when only metadata changed

**Cons:**
- More complex
- Still requires reading file content when metadata differs

## Plan

1. Implement Approach 1 first - covers 99% of use cases
2. Evaluate if hash verification (Approach 2) is needed based on real usage
