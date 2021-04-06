package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

type mmaEncoder struct {
	writer      *multipart.Writer
	attachments []MIMEMultipartAttachment
}

func newMmaEncoder(w io.Writer, attachments []MIMEMultipartAttachment) *mmaEncoder {
	return &mmaEncoder{
		writer:      multipart.NewWriter(w),
		attachments: attachments,
	}
}

func (e *mmaEncoder) Encode(v interface{}) error {
	var err error
	var soapPartWriter io.Writer

	// 1. write SOAP envelope part
	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Type", `text/xml;charset=UTF-8`)
	headers.Set("Content-Transfer-Encoding", "8bit")
	headers.Set("Content-ID", "<soaprequest@gowsdl.lib>")
	if soapPartWriter, err = e.writer.CreatePart(headers); err != nil {
		return err
	}
	xmlEncoder := xml.NewEncoder(soapPartWriter)
	if err := xmlEncoder.Encode(v); err != nil {
		return err
	}

	// 2. write attachments parts
	for _, attachment := range e.attachments {
		attHeader := make(textproto.MIMEHeader)
		attHeader.Set("Content-Type", fmt.Sprintf("application/octet-stream; name=%s", attachment.Name))
		attHeader.Set("Content-Transfer-Encoding", "binary")
		attHeader.Set("Content-ID", fmt.Sprintf("<%s>", attachment.Name))
		attHeader.Set("Content-Disposition",
			fmt.Sprintf("attachment; name=\"%s\"; filename=\"%s\"", attachment.Name, attachment.Name))
		var fw io.Writer
		fw, err := e.writer.CreatePart(attHeader)
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, bytes.NewReader(attachment.Data))
		if err != nil {
			return err
		}
	}
	// close the writer
	e.writer.Close()

	return nil
}

func (e *mmaEncoder) Flush() error {
	return e.writer.Close()
}

func (e *mmaEncoder) Boundary() string {
	return e.writer.Boundary()
}
