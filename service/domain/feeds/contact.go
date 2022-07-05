package feeds

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ContactToSave struct {
	who refs.Identity
	msg content.Contact
}

func NewContactToSave(who refs.Identity, msg content.Contact) ContactToSave {
	return ContactToSave{
		who: who,
		msg: msg,
	}
}

func (c ContactToSave) Who() refs.Identity {
	return c.who
}

func (c ContactToSave) Msg() content.Contact {
	return c.msg
}
