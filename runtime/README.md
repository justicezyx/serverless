# Runtime

```
python3 runtime.py --file=runtime_beta.py --class_name=RuntimeBeta
curl -X POST http://localhost:8080/forward \
    -H "Content-Type: application/json" \
    -d '{"prompt": "What should I do today?"}'
```
