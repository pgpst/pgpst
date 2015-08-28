package smtpd

// Wrapper is used for surrounding whole calls with defer etc.
type Wrapper interface {
	Wrap(func()) func()
}

type WrapperFunc func(func()) func()

func (w WrapperFunc) Wrap(x func()) func() {
	return w(x)
}

// Sender allows data preloading whenever a new envelope is created.
type Sender interface {
	HandleSender(func(conn *Connection)) func(conn *Connection)
}

type SenderFunc func(func(conn *Connection)) func(conn *Connection)

func (s SenderFunc) HandleSender(x func(conn *Connection)) func(conn *Connection) {
	return s(x)
}

// Recipient is called whenever client sents a RCPT command to the server.
type Recipient interface {
	HandleRecipient(func(conn *Connection)) func(conn *Connection)
}

type RecipientFunc func(func(conn *Connection)) func(conn *Connection)

func (r RecipientFunc) HandleRecipient(x func(conn *Connection)) func(conn *Connection) {
	return r(x)
}

// Delivery is called after the whole MIME message is sent to the server.
type Delivery interface {
	HandleDelivery(func(conn *Connection)) func(conn *Connection)
}

type DeliveryFunc func(func(conn *Connection)) func(conn *Connection)

func (d DeliveryFunc) HandleDelivery(x func(conn *Connection)) func(conn *Connection) {
	return d(x)
}
