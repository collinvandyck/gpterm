# gpterm (GPT terminal)

gpterm is a terminal-based integration with the OpenAI chat API.

# Example

https://user-images.githubusercontent.com/596076/230783174-67f512d2-9c8f-4131-a043-db7083b16a04.mp4

gpterm also renders Markdown and syntax-aware code blocks:

<img width="1800" alt="image-2023-05-24-T5sJJ5vm@2x" src="https://github.com/collinvandyck/gpterm/assets/596076/a33c4470-0275-4927-b346-00a68bdcc4e9">

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

