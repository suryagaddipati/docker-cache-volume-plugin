PLUGIN_NAME = suryagaddipati/cachedriver
PLUGIN_TAG ?= next

all: clean rootfs create

clean:
	@echo "### rm ./plugin"
	@rm -rf ./plugin

rootfs:
	@echo "### docker build: rootfs image with docker-cachedriver"
	@docker build -q -t ${PLUGIN_NAME}:rootfs .
	@echo "### create rootfs directory in ./plugin/rootfs"
	@mkdir -p ./plugin/rootfs
	@docker create --name tmp ${PLUGIN_NAME}:rootfs
	@docker export tmp > plugin.tar
	@gnutar xvC ./plugin/rootfs/ -f plugin.tar || true
	@echo "### copy config.json to ./plugin/"
	@cp config.json ./plugin/
	@docker rm -vf tmp	

create:
	@echo "### remove existing plugin ${PLUGIN_NAME} if exists"
	@docker plugin rm -f ${PLUGIN_NAME} || true
	@echo "### create new plugin ${PLUGIN_NAME} from ./plugin"
	@docker plugin create ${PLUGIN_NAME} ./plugin

enable:		
	@echo "### enable plugin ${PLUGIN_NAME}"		
	@docker plugin enable ${PLUGIN_NAME}

push:  clean rootfs create enable
	@echo "### push plugin ${PLUGIN_NAME}"
	@docker plugin push ${PLUGIN_NAME}


