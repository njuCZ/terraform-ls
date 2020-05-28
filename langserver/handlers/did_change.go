package handlers

import (
	"context"
	"encoding/json"

	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	lsp "github.com/sourcegraph/go-lsp"
)

func (h *logHandler) TextDocumentDidChange(ctx context.Context, params DidChangeTextDocumentParams) error {
	p := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: params.TextDocument.URI,
			},
			Version: params.TextDocument.Version,
		},
		ContentChanges: params.ContentChanges,
	}

	fs, err := lsctx.Filesystem(ctx)
	if err != nil {
		return err
	}

	fh := ilsp.VersionedFileHandler(p.TextDocument)
	f, err := fs.GetFile(fh)
	if err != nil {
		return err
	}

	// old version change, just skip
	if p.TextDocument.Version <= f.Version() {
		h.logger.Printf("skip old version %d change: %#v, current file version %d", p.TextDocument.Version, params, f.Version())
		return nil
	}

	// missing version
	if p.TextDocument.Version > f.Version()+1 {
		h.logger.Printf("missing version %d change: %#v, current file version %d", p.TextDocument.Version, params, f.Version())
		return nil
	}

	changes, err := ilsp.FileChanges(params.ContentChanges, f)
	if err != nil {
		return err
	}
	return fs.Change(fh, changes)
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier      `json:"textDocument"`
	ContentChanges []lsp.TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI lsp.DocumentURI `json:"uri"`
	/**
	 * The version number of this document.
	 */
	Version int `json:"version"`
}

// UnmarshalJSON implements non-strict json.Unmarshaler.
func (v *DidChangeTextDocumentParams) UnmarshalJSON(b []byte) error {
	type t DidChangeTextDocumentParams
	return json.Unmarshal(b, (*t)(v))
}

// UnmarshalJSON implements non-strict json.Unmarshaler.
func (v *VersionedTextDocumentIdentifier) UnmarshalJSON(b []byte) error {
	type t VersionedTextDocumentIdentifier
	return json.Unmarshal(b, (*t)(v))
}
