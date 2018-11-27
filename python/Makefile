.PHONY: install build clean cleanall

# See https://grpc.io/docs/quickstart/python.html.
install:
	python -m pip install grpcio --user
	python -m pip install grpcio-tools googleapis-common-protos --user

build:
	python -m grpc_tools.protoc -I../ --python_out=. --grpc_python_out=. ../kv.proto

clean:
	rm -f *.pyc

cleanall: clean
	rm -f kv_pb2_grpc.py kv_pb2.py