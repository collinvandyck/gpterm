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

# Usage Stats

While OpenAPI is incredibly cheap at the moment, it's not totally free. You can view your usage at any time:

	❯ gpterm usage

	Prompt:     58761
	Completion: 19722
	Total:      78483
	Cost:       $0.16 ($0.002 per 1K tokens)
	
# Contributing

It's too early to accept pull requests given how early this project is. Please feel free to file an issue if you
found a bug or have a feature request.

