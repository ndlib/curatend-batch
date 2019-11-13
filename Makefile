
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
			tasks/access-to-relsext=/opt/batchs/tasks/access-to-relsext \
			tasks/add-to-queue.rb=/opt/batchs/tasks/add-to-queue.rb \
			tasks/arrange-files.rb=/opt/batchs/tasks/arrange-files.rb \
			tasks/assign-pids=/opt/batchs/tasks/assign-pids \
			tasks/characterize=/opt/batchs/tasks/characterize \
			tasks/compare-rof=/opt/batchs/tasks/compare-rof \
			tasks/conf.example=/opt/batchs/tasks/conf.example \
			tasks/csv-to-mets.rb=/opt/batchs/tasks/csv-to-mets.rb \
			tasks/csv-to-rof=/opt/batchs/tasks/csv-to-rof \
			tasks/date-stamp=/opt/batchs/tasks/date-stamp \
			tasks/do-fedora-only-content=/opt/batchs/tasks/do-fedora-only-content \
			tasks/do-fedora-only-generic-files=/opt/batchs/tasks/do-fedora-only-generic-files \
			tasks/do-fedora-only-works=/opt/batchs/tasks/do-fedora-only-works \
			tasks/env=/opt/batchs/tasks/env \
			tasks/error=/opt/batchs/tasks/error \
			tasks/fedora-to-rof=/opt/batchs/tasks/fedora-to-rof \
			tasks/file-to-url=/opt/batchs/tasks/file-to-url \
			tasks/filename-normalize=/opt/batchs/tasks/filename-normalize \
			tasks/get-from-bendo=/opt/batchs/tasks/get-from-bendo \
			tasks/get-from-osf=/opt/batchs/tasks/get-from-osf \
			tasks/iiif_template.erb=/opt/batchs/tasks/iiif_template.erb \
			tasks/index=/opt/batchs/tasks/index \
			tasks/ingest=/opt/batchs/tasks/ingest \
			tasks/jsonld-to-rof=/opt/batchs/tasks/jsonld-to-rof \
			tasks/move-files-for-bendo=/opt/batchs/tasks/move-files-for-bendo \
			tasks/osf-to-rof=/opt/batchs/tasks/osf-to-rof \
			tasks/reingest-rof=/opt/batchs/tasks/reingest-rof \
			tasks/remove-blacklisted-rofs=/opt/batchs/tasks/remove-blacklisted-rofs \
			tasks/remove-csv-bom=/opt/batchs/tasks/remove-csv-bom \
			tasks/rof-to-csv=/opt/batchs/tasks/rof-to-csv \
			tasks/rof-to-mellon=/opt/batchs/tasks/rof-to-mellon \
			tasks/start=/opt/batchs/tasks/start \
			tasks/start-fedora-only=/opt/batchs/tasks/start-fedora-only \
			tasks/start-osf-archive-ingest=/opt/batchs/tasks/start-osf-archive-ingest \
			tasks/start-reingest=/opt/batchs/tasks/start-reingest \
			tasks/upload-to-bendo=/opt/batchs/tasks/upload-to-bendo \
			tasks/upload-to-googledrive.rb=/opt/batchs/tasks/upload-to-googledrive.rb \
			tasks/validate=/opt/batchs/tasks/validate \
			tasks/work-xlat=/opt/batchs/tasks/work-xlat
