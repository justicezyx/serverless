import time

class RuntimeAlpha:
    def load(self):
        # This function needs to be called every time the container starts.
        time.sleep(4)

    def generate(self, args):
        # This is just a mock function to simulate processing of the user request.
        time.sleep(0.75)
        prompt = args["prompt"]
        return f"Given your question: {prompt}. I think the best answer is to buy ice cream."
