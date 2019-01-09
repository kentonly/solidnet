
flags = -gcflags=all="-N -l"
bin_dir = bin

all:
	@if test ! -d $(bin_dir) ; \
    then \
		mkdir $(bin_dir); \
    fi; \

	go build $(flags) -o server server.go handler.go packet.go && mv server $(bin_dir) -f
	go build $(flags) -o client client.go packet.go && mv client $(bin_dir) -f

clean:
	rm $(bin_dir) -rf
	
