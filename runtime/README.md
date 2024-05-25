# Runtime

```
python3 runtime.py --file=runtime_beta.py --class_name=RuntimeBeta
curl -X POST http://localhost:5000/invoke \
    -H "Content-Type: application/json" \
    -d '{"args": {"prompt": "What should I do today?"}}'
```
