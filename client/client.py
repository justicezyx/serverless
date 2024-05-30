import requests
import argparse
import threading

base_url = "http://localhost:8080/"
headers = {
    "Content-Type": "application/json"
}
payload = {
    "args": {
        "prompt": "What should I do today?"
    }
}

class Alpha:
    def __init__(self, args, user):
        self.url = base_url + "alpha"
        self.headers = headers
        self.headers["User"] = user
        self.payload = {"args": args}

    def __call__(self):
        return requests.post(self.url, json=self.payload, headers=self.headers)

class Beta:
    def __init__(self, args, user):
        self.url = base_url + "beta"
        self.headers = headers
        self.headers["User"] = user
        self.payload = {"args": args}

    def __call__(self):
        return requests.post(self.url, json=self.payload, headers=self.headers)

def LoopAlpha():
    while True:
        try:
            response = Alpha({"prompt": "What should I do today?"}, args.user)()
            print("Alpha response:", response.text)
        except Exception as e:
            print("Request failed:", e)

def LoopBeta():
    while True:
        try:
            response = Beta({"prompt": "What should I do today?"}, args.user)()
            print("Beta response:", response.content)
        except Exception as e:
            print("Request failed:", e)

# Usage example
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Invoke Alpha or Beta Runtime with a user.')
    parser.add_argument('--user', required=True, help='User identifier for the request')
    parser.add_argument('--concurrency', default=1, help='How many concurrent' +
                        'calls to each function')
    args = parser.parse_args()

    threads = []
    for i in range(args.concurrency):
        alphaThread = threading.Thread(target=LoopAlpha)
        alphaThread.start()
        threads.append(alphaThread)

        betaThread = threading.Thread(target=LoopBeta)
        betaThread.start()
        threads.append(betaThread)

    # Wait for all threads to complete
    for thread in threads:
        thread.join()
