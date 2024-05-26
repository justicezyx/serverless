import importlib.util
import argparse
from flask import Flask, request, jsonify

app = Flask(__name__)

def load_class_from_file(file_path, class_name):
    spec = importlib.util.spec_from_file_location(class_name, file_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    clazz = getattr(module, class_name)
    return clazz

class ServerlessRuntime:
    def __init__(self, class_path, class_name):
        self.class_path = class_path
        self.class_name = class_name
        self.runtime_class = load_class_from_file(class_path, class_name)()
        # Ensure load function is called before serving requests.
        # This is required in the assignment spec.
        self.runtime_class.load()

    def handle_request(self, args):
        return self.runtime_class.generate(args)

runtime_instance = None

@app.route('/invoke', methods=['POST'])
def invoke():
    data = request.json
    args = data.get('args', {})

    response = runtime_instance.handle_request(args)

    return jsonify({"response": response})

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Serverless Runtime')
    parser.add_argument('--file', required=True, help='Path to the Python file')
    parser.add_argument('--class_name', required=True, help='Name of the class to load')

    args = parser.parse_args()

    runtime_file_path = args.file
    runtime_class_name = args.class_name

    # Initializing a runtime instance with the provided file path and class name
    runtime_instance = ServerlessRuntime(runtime_file_path, runtime_class_name)

    # Use host='0.0.0.0' to bind to all local IP address.
    # This seems necessary when running inside docker container.
    app.run(host='0.0.0.0', port=5000)
