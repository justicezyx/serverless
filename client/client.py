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
        try:
            response = requests.post(self.url, json=self.payload, headers=self.headers)
            print(response.content)
            response.raise_for_status()  # Raise an error for bad status codes
            try:
                return response.json()
            except ValueError:
                print("Response content is not valid JSON:", response.text)
                return None
        except requests.exceptions.RequestException as e:
            print("Request failed:", e)
            return None

class Beta:
    def __init__(self, args, user):
        self.url = base_url + "beta"
        self.headers = headers
        self.headers["User"] = user
        self.payload = {"args": args}

    def __call__(self):
        try:
            response = requests.post(self.url, json=self.payload, headers=self.headers)
            response.raise_for_status()  # Raise an error for bad status codes
            try:
                return response.json()
            except ValueError:
                print("Response content is not valid JSON:", response.text)
                return None
        except requests.exceptions.RequestException as e:
            print("Request failed:", e, response.body)
            return None

def LoopAlpha():
    while True:
        # Invoke the Alpha Runtime
        response = Alpha({"prompt": "What should I do today?"}, args.user)()
        print("Alpha response:", response)

def LoopBeta():
    while True:
        response = Beta({"prompt": "What should I do today?"}, args.user)()
        print("Beta response:", response)

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
