# Runtime

```
python3 runtime.py --file=runtime_beta.py --class_name=RuntimeBeta
curl -X POST http://127.0.0.1:5000/invoke \
    -H "Content-Type: application/json" \
    -d '{"args": {"prompt": "What should I do today?"}}'
```
