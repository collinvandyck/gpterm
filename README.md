# gpterm (GPT terminal)

gpterm is a terminal-based integration with the OpenAI chat API.

# Demo

https://user-images.githubusercontent.com/596076/229314588-98bf28df-56bb-48d3-921a-095300381416.mp4

# Installation

Requirements:

- Go 1.18 or higher
- GOPATH env variable set (install will write to `$GOPATH/bin`)
- OpenAI API Key

To install from source:

	git clone https://github.com/collinvandyck/gpterm.git
	cd gpterm
	make install

# Getting Started

Because it uses the OpenAI API, an API key is required before you can start:

	# set api key
	gpterm auth

Once the API key has been set, have fun!

	# enter an interactive session
	gpterm

Note that there are other subcommand attached to the `gpterm` command. Those are meant for development and are
not useful outside of the repo.

# Storage

Chat history and your API key are stored in a sqlite database in:

	~/.config/gpterm

# Upcoming

## Configurable Roles

Users will be able to instruct OpenAI how to act. Currently it's hardcoded to `You are a helpful assistant.`

Additionally, users should be able to change the name of the assistant as rendered in the chatlog. Current it is `ChatGPT`.

## Gist Support

Conversations should be able to be uploaded to a gist.

## Hotkey Support

Actions should be able to be found to user-configured keys.

# Contributing

It's too early to accept pull requests given how early this project is. Please feel free to file an issue if you
found a bug or have a feature request.

