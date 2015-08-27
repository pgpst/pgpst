package mailer

import (
	"bufio"
	"bytes"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/pgpst/pgpst/pkg/models"
)

func analyzeEmail(n *models.EmailNode, input []byte) error {
	// Figure out if newlines are \n or \r\n
	var nl []byte
	firstNewline := bytes.Index(input, []byte("\n"))
	if firstNewline-1 >= 0 && input[firstNewline-1] == '\r' {
		nl = []byte("\r\n")
	} else {
		nl = []byte("\n")
	}
	doubleNl := append(nl, nl...)

	// Parse the header
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(input)))
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return err
	}
	n.Headers = mail.Header(headers)

	// Calculate where the parts are
	seperator := bytes.Index(input, doubleNl)
	n.HeaderPosition = [2]int{n.BasePosition, n.BasePosition + seperator + len(nl)}
	n.BodyPosition = [2]int{n.BasePosition + seperator + len(nl), n.BasePosition + len(input)}

	// Check if the node is multipart
	media, params, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(media, "multipart/") {
		return nil
	}
	if _, ok := params["boundary"]; !ok {
		return nil
	}

	// Prepare children slice
	n.Children = []*models.EmailNode{}

	// Prepare byte slices for comparsion
	var (
		dashBoundaryDashNl = []byte("--" + params["boundary"] + "--" + string(nl))
		dashBoundaryDash   = dashBoundaryDashNl[:len(dashBoundaryDashNl)-2]
		dashBoundaryNl     = []byte("--" + params["boundary"] + string(nl))
	)

	// Analyze where the multipart body starts
	start := bytes.Index(input, dashBoundaryNl)
	end := bytes.Index(input, dashBoundaryDashNl)
	if end == -1 {
		end = bytes.Index(input, dashBoundaryDash)
	}

	// Trim the data to only get the body without the wrappers
	bodyBaseIndex := n.BasePosition + start + len(dashBoundaryNl)
	partsData := input[start+len(dashBoundaryNl) : end]

	// Split the parts into multipart body parts
	for _, part := range bytes.Split(partsData, dashBoundaryNl) {
		np := &models.EmailNode{
			BasePosition: bodyBaseIndex,
		}

		if err := analyzeEmail(np, part); err != nil {
			return err
		}

		bodyBaseIndex += len(part) + len(dashBoundaryNl)

		n.Children = append(n.Children, np)
	}

	return nil
}

func findEncrypted(node *models.EmailNode) bool {
	media, _, _ := mime.ParseMediaType(node.Headers.Get("Content-Type"))
	if media == "multipart/encrypted" || media == "application/pgp-encrypted" {
		return true
	}

	if strings.HasPrefix(media, "multipart/") {
		for _, node := range node.Children {
			if findEncrypted(node) {
				return true
			}
		}
	}

	return false
}
