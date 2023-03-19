# gpterm (GPT terminal)

gpterm is a terminal-based integraiton with the OpenAI chat APIs.

Chat history and other metadata are stored in

	~/.config/gpterm

# Installation

Requirements:

- Go 1.18 or higher
- OpenAPI API Key

To install from source:

	git clone https://github.com/collinvandyck/gpterm.git
	cd gpterm
	make install

# Getting Started

Because it uses the OpenAPI API, and API key is required before you can start:

	# set api key
	gpterm auth

Once the API key has been set, have fun!

	# enter an interactive session
	gpterm
	
# Contributing

It's too early to accept pull requests given how early this project is. Please feel free to file an issue if you
found a bug or have a feature request.

