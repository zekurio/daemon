package quickembed

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// QuickEmbed is a helper struct for quickly creating embeds.
type QuickEmbed struct {
	emb *discordgo.MessageEmbed
}

// New creates a new QuickEmbed instance.
func New() *QuickEmbed {
	return &QuickEmbed{
		emb: &discordgo.MessageEmbed{},
	}
}

func (q *QuickEmbed) SetTitle(title string) *QuickEmbed {
	q.emb.Title = title
	return q
}

func (q *QuickEmbed) SetDescription(description string) *QuickEmbed {
	q.emb.Description = description
	return q
}

func (q *QuickEmbed) SetURL(url string) *QuickEmbed {
	q.emb.URL = url
	return q
}

func (q *QuickEmbed) SetColor(color int) *QuickEmbed {
	q.emb.Color = color
	return q
}

func (q *QuickEmbed) SetAuthor(name, iconURL, pURL, url string) *QuickEmbed {
	q.emb.Author = &discordgo.MessageEmbedAuthor{
		Name:         name,
		IconURL:      iconURL,
		URL:          url,
		ProxyIconURL: pURL,
	}
	return q
}

func (q *QuickEmbed) SetThumbnail(url, pURL string, width, height int) *QuickEmbed {
	q.emb.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL:      url,
		ProxyURL: pURL,
		Width:    width,
		Height:   height,
	}
	return q
}

func (q *QuickEmbed) SetImage(url, pURL string, width, height int) *QuickEmbed {
	q.emb.Image = &discordgo.MessageEmbedImage{
		URL:      url,
		ProxyURL: pURL,
		Width:    width,
		Height:   height,
	}
	return q
}

func (q *QuickEmbed) SetFooter(text, iconURL, pURL string) *QuickEmbed {
	q.emb.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return q
}

func (q *QuickEmbed) SetTimestamp(timestamp ...time.Time) *QuickEmbed {

	if len(timestamp) > 0 {
		q.emb.Timestamp = timestamp[0].Format("02/01/2006 15:04:05")
	} else {
		q.emb.Timestamp = time.Now().Format("02/01/2006 15:04:05")
	}

	return q
}

func (q *QuickEmbed) Build() *discordgo.MessageEmbed {
	return q.emb
}

func (q *QuickEmbed) AddField(name, value string, inline ...bool) *QuickEmbed {

	if value == "" {
		value = "`nil`"
	}

	q.emb.Fields = append(q.emb.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: len(inline) > 0 && inline[0],
	})
	return q
}
