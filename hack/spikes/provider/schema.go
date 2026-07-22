package provider

import "github.com/PlatformRelay/assent/schemas"

// request.schema.json/response.schema.json were promoted unchanged to
// schemas/provider/v1alpha1/ (REQ-P3-E1-S03-01). Go's //go:embed forbids
// ".." patterns (embed.FS can't address a parent directory), so this
// throwaway spike package delegates to the promoted schemas package's
// exported validators instead of embedding the files itself — the public
// ValidateRequest/ValidateResponse surface this package's tests already
// depend on is unchanged.

// ValidateRequest checks raw JSON against the promoted request.schema.json.
func ValidateRequest(raw []byte) error { return schemas.ValidateProviderRequest(raw) }

// ValidateResponse checks raw JSON against the promoted response.schema.json.
func ValidateResponse(raw []byte) error { return schemas.ValidateProviderResponse(raw) }
