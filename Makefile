AGENT_APP=agentd
COMMANDER_APP=commanderd

NOHUP_HELPER=firespotter-nohup.sh
NOHUP_HELPER_AGENT_ONLY=firespotter-nohup.sh
COMMANDER_DEFAULT_CONF=commanderd.conf.json

RELEASE_ARCHIVE=firespotter.tar.gz


release: build
	echo "Creating a release..."
	tar --create --file=$(RELEASE_ARCHIVE) $(AGENT_APP) $(COMMANDER_APP) $(NOHUP_HELPER) $(NOHUP_HELPER_AGENT_ONLY) $(COMMANDER_DEFAULT_CONF)
	echo "-> Release packed, OK!"
	echo "Unpack: 'tar -xf $(RELEASE_ARCHIVE)'"

extract_release:
	tar -xf $(RELEASE_ARCHIVE)

build: clean
	echo "Building the agent app..."
	go build -ldflags="-s -w" -a -o $(AGENT_APP) agent/main.go
	echo "Building the commander app..."
	go build -ldflags="-s -w" -a -o $(COMMANDER_APP) commander/main.go
	echo "-> Build finished, OK!"

clean:
	rm -f $(AGENT_APP)
	rm -f $(COMMANDER_APP)
	rm -f $(RELEASE_ARCHIVE)

.SILENT: release build clean extract_release
