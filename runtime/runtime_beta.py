class RuntimeBeta:
    def load(self):
        # This function needs to be called every time the container starts.
        import time
        time.sleep(2.4)

    def generate(self, args):
        # This is just a mock function to simulate processing of the user request.
        import time
        time.sleep(1.75)
        prompt = args["prompt"]
        return f"Given your question: {prompt}. I think the best answer is to get a hamburger."
