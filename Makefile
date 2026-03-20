BINARY := clipu
PREFIX ?= ~/.local

.PHONY: build install uninstall clean

build:
	go build -o $(BINARY) .

install: build
	install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
	rm -f __debug_bin*
