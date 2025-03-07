package auth

import (
	"context"
	"fmt"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/shared/utils/arrays"
	"github.com/golang-jwt/jwt/v4"
	"github.com/openshift-online/ocm-sdk-go/authentication"
)

// Context key type defined to avoid collisions in other pkgs using context
// See https://golang.org/pkg/context/#WithValue
type contextKey string

const (
	// Context Keys
	// FilterByOrganisation is used to determine whether resources are filtered by a user's organisation or as an individual owner
	contextFilterByOrganisation contextKey = "filter-by-organisation"
	contextIsAdmin              contextKey = "is_admin"
)

var (
	// ocm token claim keys
	tenantUsernameClaim string = "username"
	tenantIdClaim       string = "org_id"
	tenantOrgAdminClaim string = "is_org_admin" // same key used in mas-sso tokens

	// sso.redhat.com token claim keys
	alternateTenantUsernameClaim string = "preferred_username" // same key used in mas-sso tokens
	tenantUserIdClaim            string = "account_id"

	// mas-sso token claim keys
	// NOTE: This should be removed once we migrate to sso.redhat.com as it will no longer be needed (TODO: to be removed as part of MGDSTRM-6159)
	alternateTenantIdClaim = "rh-org-id"
)

func GetUsernameFromClaims(claims jwt.MapClaims) string {
	if idx, val := arrays.FindFirst(func(x interface{}) bool { return x != nil }, claims[tenantUsernameClaim], claims[alternateTenantUsernameClaim]); idx != -1 {
		return val.(string)
	}
	return ""
}

func GetAccountIdFromClaims(claims jwt.MapClaims) string {
	if claims[tenantUserIdClaim] != nil {
		return claims[tenantUserIdClaim].(string)
	}
	return ""
}

func GetOrgIdFromClaims(claims jwt.MapClaims) string {
	if claims[tenantIdClaim] != nil {
		if orgId, ok := claims[tenantIdClaim].(string); ok {
			return orgId
		}
	}

	// NOTE: This should be removed once we migrate to sso.redhat.com as it will no longer be needed (TODO: to be removed as part of MGDSTRM-6159)
	if claims[alternateTenantIdClaim] != nil {
		if orgId, ok := claims[alternateTenantIdClaim].(string); ok {
			return orgId
		}
	}

	return ""
}

func GetIsOrgAdminFromClaims(claims jwt.MapClaims) bool {
	if claims[tenantOrgAdminClaim] != nil {
		return claims[tenantOrgAdminClaim].(bool)
	}
	return false
}

func GetIsAdminFromContext(ctx context.Context) bool {
	isAdmin := ctx.Value(contextIsAdmin)
	if isAdmin == nil {
		return false
	}
	return isAdmin.(bool)
}

func SetFilterByOrganisationContext(ctx context.Context, filterByOrganisation bool) context.Context {
	return context.WithValue(ctx, contextFilterByOrganisation, filterByOrganisation)
}

func SetIsAdminContext(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, contextIsAdmin, isAdmin)
}

func GetFilterByOrganisationFromContext(ctx context.Context) bool {
	filterByOrganisation := ctx.Value(contextFilterByOrganisation)
	if filterByOrganisation == nil {
		return false
	}
	return filterByOrganisation.(bool)
}

func SetTokenInContext(ctx context.Context, token *jwt.Token) context.Context {
	return authentication.ContextWithToken(ctx, token)
}

func GetClaimsFromContext(ctx context.Context) (jwt.MapClaims, error) {
	var claims jwt.MapClaims
	token, err := authentication.TokenFromContext(ctx)
	if err != nil {
		return claims, fmt.Errorf("failed to get jwt token from context: %v", err)
	}

	if token != nil && token.Claims != nil {
		claims = token.Claims.(jwt.MapClaims)
	}
	return claims, nil
}
