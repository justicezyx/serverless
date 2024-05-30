# Runtime

Runtime is implemented as docker container, and HTTP services.

To start runtime without using docker container, and curling it:
```shell
python3 runtime.py --file=runtime_beta.py --class_name=RuntimeBeta
curl -X POST http://127.0.0.1:5000/invoke \
    -H "Content-Type: application/json" \
    -d '{"args": {"prompt": "What should I do today?"}}'
```

To build and run docker image:
```shell
docker build -t runtime .
docker run -p 8001:5000 -d runtime \
    python runtime.py --file runtime_beta.py --class_name RuntimeBeta
docker run -p 8002:5000 -d runtime \
    python runtime.py --file runtime_alpha.py --class_name RuntimeAlpha
```
