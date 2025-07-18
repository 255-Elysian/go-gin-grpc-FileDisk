IDL_PATH = $(shell pwd)/idl

.PHONY: proto
proto:
	@for file in $(IDL_PATH)/*.proto; do \
		protoc -I $(IDL_PATH) $$file --go-grpc_out=$(IDL_PATH)/pb --go_out=$(IDL_PATH)/pb; \
	done
	@for file in $(shell find $(IDL_PATH)/pb/* -type f); do \
  		protoc-go-inject-tag -input=$$file; \
  	done