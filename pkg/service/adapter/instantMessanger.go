package adapter

import discord "github.com/bwmarrin/discordgo"

type InstantMessanger struct {
	session *discord.Session
}

func NewInstantMessanger(session *discord.Session) InstantMessanger {
	return InstantMessanger{
		session: session,
	}
}

func (i *InstantMessanger) SendMessage(message string) error {

}
