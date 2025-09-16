# 8fs Vision: AI-Native Edge Storage

## 🌍 Our Vision

**8fs is the first S3-compatible storage server with built-in vector storage for AI developers.**

Not a MinIO clone—8fs unifies object storage and vector embeddings in a lightweight binary for local AI labs, from laptops to Raspberry Pi clusters.

> *"S3 + vector storage in one <50MB binary—perfect for indie AI workflows."*

## 🚨 **Current Status: MVP Focus (Weeks 1-2)**
We're all-in on **v0.1**: A working S3 + vector storage API, shipping in 2 weeks. No fluff, no overpromises—P2P sync, WASM, and UI are delayed until the core proves itself.

**Reality Check**: This is a day-old repo (https://github.com/8fs-io/core). Success means shipping fast, backing claims with benchmarks, and building a community—not dreaming big and fizzling out.

---

## 🎯 The Problem We're Solving

### Pain Points:
- **MinIO/Ceph**: 100MB+ binaries, too heavy for Raspberry Pi or local AI labs
- **Vector Databases**: Separate infra (Pinecone, Weaviate) adds complexity  
- **Local AI Workflows**: Need S3 for models and vector search for embeddings
- **Edge Deployments**: Most solutions demand 200MB+ RAM, choking small devices

### Market Gap:
The local AI boom (Ollama, LLaMA.cpp, ComfyUI) demands simple, unified storage for models and embeddings without enterprise bloat.

---

## 🚀 Our Solution: **S3 + Vectors in One Binary**

### **MVP Value Proposition (v0.1 - Shipping in 2 Weeks):**
> **"S3-compatible storage with basic vector search—proving unified AI storage works."**

### **What We're Shipping:**

#### 1. **Minimal Vector Storage**
- **Endpoints**: `/vectors/embeddings` (POST), `/vectors/search` (POST)
- **SQLite storage**, 1K vector limit (384-dim for MVP)
- **Basic cosine similarity** search (HNSW in v0.2)
- **Binary size**: <50MB, tested on Raspberry Pi

#### 2. **S3 Compatibility**  
- **Existing S3 API** (buckets, objects) fully functional
- **Works with AWS CLI, boto3, s3fs** out of the box
- **Unified storage**: S3 for models, vectors for embeddings
- **Python examples** in README for AI workflows

### **What's NOT in v0.1:**
- ❌ P2P synchronization (v0.3)
- ❌ WASM browser build (v0.2)  
- ❌ Web UI (v0.2)
- ❌ Advanced vector search (HNSW, batch ops)
- ❌ Multi-node clustering or enterprise features

### **Why 8fs?**

| Feature | 8fs (v0.1) | MinIO | Pinecone | Weaviate | AWS S3 |
|---------|-------------|--------|----------|----------|--------|
| **Binary Size** | <50MB | 100MB+ | N/A (SaaS) | 500MB+ | N/A (Cloud) |
| **Vector Storage** | ✅ Basic | ❌ | ✅ Advanced | ✅ | ❌ |
| **S3 Compatibility** | ✅ | ✅ | ❌ | ❌ | ✅ |
| **Raspberry Pi Fit** | ✅ | ⚠️ Heavy | ❌ | ❌ | ❌ |
| **Setup Complexity** | Zero-config | Complex | SaaS | Complex | Cloud |

---

## 🎯 Target Market

### **Primary: Indie AI Developers** (0-6 Months)
- **Who**: Solo devs, AI hobbyists, homelabbers
- **Use Cases**: Local LLM experiments, RAG apps, model fine-tuning
- **Pain**: Need free, simple S3 + vector storage for local AI
- **Acquisition**: Hacker News, r/LocalLLaMA, r/selfhosted, AI Discords

### **Secondary: Edge AI Startups** (6-18 Months)
- **Who**: Startups in retail, industrial, IoT AI
- **Use Cases**: Edge inference, distributed models, data pipelines
- **Pain**: Complex infra, cloud costs, latency
- **Acquisition**: Product Hunt, AI conferences, blogs

### **Tertiary: AI Infrastructure Teams** (18+ Months)
- **Who**: Companies with large AI/ML needs
- **Use Cases**: Multi-region AI, hybrid cloud, compliance
- **Pain**: Vendor lock-in, cost, data sovereignty
- **Acquisition**: Enterprise sales, partnerships

---

## 🛠️ Technical Implementation (v0.1 MVP)

### **Vector API (1K Vectors Max)**
```bash
# Store embeddings (384 dimensions)
curl -X POST http://localhost:8080/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "embedding": [0.1, 0.2, ...], "metadata": {"type": "document"}}'

# Basic cosine similarity search
curl -X POST http://localhost:8080/vectors/search \
  -H "Content-Type: application/json" \
  -d '{"query": [0.1, 0.2, ...], "top_k": 5}'
```

### **Python Example**
```python
import requests
import numpy as np

# Store model via S3 API
with open("model.gguf", "rb") as f:
    requests.put("http://localhost:8080/my-ai-project/model.gguf", data=f)

# Store embeddings
embedding = np.random.random(384).tolist()  # Small dims for MVP
response = requests.post("http://localhost:8080/vectors/embeddings", json={
    "id": "doc1",
    "embedding": embedding,
    "metadata": {"filename": "doc1.txt"}
})

# Search similar embeddings
results = requests.post("http://localhost:8080/vectors/search", json={
    "query": embedding,
    "top_k": 5
}).json()
print(f"Found {len(results['matches'])} similar vectors")
```

### **Storage Architecture**
```
8fs-data/
├── objects/          # S3 objects (existing filesystem)
├── metadata.db       # SQLite database
│   ├── objects       # S3 metadata (existing)
│   ├── embeddings    # Vector embeddings (new)
│   └── vector_index  # Basic search index (new)
└── config.yml        # Server config (existing)
```

---

## 🎪 Future Features (Post-v0.1)

### **1. WASM Deployment (v0.2)**
Run 8fs in browsers for edge AI:
```html
<script src="https://cdn.8fs.io/8fs.wasm"></script>
<script>
  const s3 = new S3Server({ memory: "50MB" });
  await s3.start();
</script>
```

### **2. Smart Edge Clustering (v0.3)**
P2P sync with zero-config:
```yaml
edge:
  auto_discovery: true
  sync_strategy: "smart"
  conflict_resolution: "timestamp"
  bandwidth_limit: "1MB/s"
```

### **3. Web UI (v0.2)**
React-based dashboard for buckets and vectors.

### **4. AI Workflow SDK (v0.3)**
```python
import s3fs_ai
store = s3fs_ai.connect("http://localhost:8080")
store.put_model("model.gguf", auto_embed=True)
results = store.search_similar("query", top_k=10)
```

---

## 💰 Business Model (Delayed)

### **Phase 1: Pure OSS (0-12 Months)**
- **100% MIT License**: S3 + basic vectors free forever
- **Funding**: GitHub Sponsors, Buy Me a Coffee
- **Focus**: Adoption (stars, downloads), not revenue

### **Phase 2: Sustainability (12+ Months, 1K+ Stars)**
- **Free Core**: S3 + vectors stay open source
- **Hosted Option**: 8fs Cloud for zero-ops (e.g., $10/month)
- **Add-Ons**: SSO, support (only if demanded)

---

## 🏆 Competitive Positioning

| Feature | 8fs | MinIO | Pinecone | Weaviate | AWS S3 |
|---------|-----|-------|----------|----------|--------|
| **Binary Size** | <50MB | 100MB+ | N/A (SaaS) | 500MB+ | N/A (Cloud) |
| **Vector Storage** | ✅ Basic | ❌ | ✅ Advanced | ✅ | ❌ |
| **Edge Fit** | ✅ | ⚠️ Heavy | ❌ | ❌ | ❌ |
| **Offline Operation** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **S3 Compatibility** | ✅ | ✅ | ❌ | ❌ | ✅ |
| **WASM Support** | ❌ (v0.2) | ❌ | ❌ | ❌ | ❌ |
| **Cost (Self-Hosted)** | Free | Free | N/A | Free | N/A |


---

## 🎉 The Future

**8fs is the SQLite of AI storage**—simple, unified, and built for indie AI devs.

We're not chasing enterprise or cloud giants. We're the go-to for local AI labs, delivering S3 + vectors in a single, lightweight binary.

**Join us to make AI storage dead simple.** Start with vectors, grow smart. 🚀
