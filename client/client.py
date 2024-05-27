import requests
import argparse

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
            print("Request failed:", e)
            return None

# Usage example
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Invoke Alpha or Beta Runtime with a user.')
    parser.add_argument('--user', required=True, help='User identifier for the request')
    args = parser.parse_args()

    # Invoke the Alpha Runtime
    alpha_response = Alpha({"prompt": "What should I do today?"}, args.user)()
    print("Alpha response:", alpha_response)
    alpha_response = Alpha({"prompt": "What should I eat today?"}, args.user)()
    print("Alpha response:", alpha_response)
    # Invoke the Beta Runtime
    beta_response = Beta({"prompt": "What should I eat today?"}, args.user)()
    print("Beta response:", beta_response)

