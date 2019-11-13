
GOCMD:=go
VERSION:=$(shell git describe --always)
PACKAGES:=$(shell go list ./... | grep -v /vendor/)
GO15VENDOREXPERIMENT=1

.PHONY: all test clean rpm

all: curatend-batch 

curatend-batch: $(wildcard *.go)
	go build .

test:
	$(GOCMD)  test -v $(PACKAGES)

clean:
	        rm -f curatend-batch 

rpm: curatend-batch
	               fpm -t rpm -s dir \
	               --name curatend-batch \
	                --version $(VERSION) \
	                --vendor ndlib \
	                --maintainer DLT \
	                --description "batch ingest daemon" \
	                --rpm-user app \
	                --rpm-group app \
			curatend-batch=/opt/batchs/bin/curatend-batch \
			tasks/csv-to-mets.rb=/opt/batchs/tasks/csv-to-mets.rb

