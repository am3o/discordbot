# discordbot
![Go](https://github.com/Am3o/discordbot/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/am3o/discordbot)](https://goreportcard.com/report/github.com/am3o/discordbot)

The discord bot is a naiv implementation for a message bot, which can be added to your discord server and the behavior is like the slack bot. The application can execute actions with message buzzwords like `!foo` and return famouse quotes like `> bar - foo`. In the dictionary file are included some quotes or the reactions of the buzzwords.  The bot is a self-hosted application and don't need much resources, which means, e.g. you can run it on a regular raspberry pi or build your own docker image and run it everywhere.

## Register on discord

Create your own discord bot on discord. The discord wants to register your bot to create a unique bot token. So you have to create your own application. For more information read the [offical documentation](https://discord.com/developers/docs/intro)

## Installation

Compile the source or use the docker image to execute the application. There two important settings, which are needed to run the discord bot. The first one is the discord token, which is needed for the authentication between discord and bot. The second one is the path to the dictionary which includes all actions like buzzwords and quotes.
 
## Metrics

The bot provides some metrics, which can be used by prometheus/grafana. So there is a monitoring option for the service.

