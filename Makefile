ACCOUNT := pinylin
REPO := temporal-example
#TAG := latest
# IMAGE := $(ACCOUNT)/$(REPO)
SERVER := localhost:5001
IMAGE := $(REPO)
TAG := $(s)
ifdef INPUT
#	SERVER := -$(s)
	SERVER_DF := .$(s)
endif



help:                    ## Show help message
	@printf "PLEASE REMEMBER TO LOGIN INTO YOUR DOCKER REGISTRY BEFORE INVOKING ANY TARGET!!\n\n"
	@printf "Variables you can set are:\n"
	@printf "\tREGISTRY: the registry server url\n"
	@printf "\tACCOUNT: the user/organization to be used\n"
	@printf "\tREPO: the repository/image inside the registry\n"
	@printf "\tTAG: the tag of the image\n"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/##//'

build:                   ## Build the image
	@docker build -t $(SERVER)/$(IMAGE):$(TAG) -f Dockerfile$(SERVER_DF) .

push:                    ## Push the image
	@docker push $(SERVER)/$(IMAGE):$(TAG)