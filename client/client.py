import requests

base_url = "http://localhost:8080/"
headers = {
    "Content-Type": "application/json"
}
data = {
    "args": {
        "prompt": "What should I do today?"
    }
}

class Alpha:
    def __init__(self, payload):
        self.url = base_url + "alpha"

    def __call__(self):
        try:
            response = requests.post(self.url, json=data, headers=headers)
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
    def __init__(self, payload):
        self.url = base_url + "beta"

    def __call__(self):
        try:
            response = requests.post(self.url, json=data, headers=headers)
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
    # Invoke the Alpha Runtime
    alpha_response = Alpha({"prompt": "What should I eat today?"})()
    print("Alpha response:", alpha_response)
    alpha_response = Alpha({"prompt": "What should I eat today?"})()
    print("Alpha response:", alpha_response)

    # Invoke the Beta Runtime
    beta_response = Beta({"prompt": "What should I eat today?"})()
    print("Beta response:", beta_response)

