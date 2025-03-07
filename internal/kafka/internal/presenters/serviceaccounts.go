package presenters

import (
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/api/public"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
)

func ConvertServiceAccountRequest(account public.ServiceAccountRequest) *api.ServiceAccountRequest {
	return &api.ServiceAccountRequest{
		Name:        account.Name,
		Description: account.Description,
	}
}

func PresentServiceAccount(account *api.ServiceAccount) *public.ServiceAccount {
	reference := PresentReference(account.ID, account)
	return &public.ServiceAccount{
		ClientId:        account.ClientID,
		ClientSecret:    account.ClientSecret,
		Name:            account.Name,
		Description:     account.Description,
		DeprecatedOwner: account.Owner,
		CreatedAt:       account.CreatedAt,
		CreatedBy:       account.Owner,
		Id:              reference.Id,
		Kind:            reference.Kind,
		Href:            reference.Href,
	}
}

func PresentServiceAccountListItem(account *api.ServiceAccount) public.ServiceAccountListItem {
	ref := PresentReference(account.ID, account)
	return public.ServiceAccountListItem{
		Id:              ref.Id,
		Kind:            ref.Kind,
		Href:            ref.Href,
		ClientId:        account.ClientID,
		Name:            account.Name,
		DeprecatedOwner: account.Owner,
		Description:     account.Description,
		CreatedAt:       account.CreatedAt,
		CreatedBy:       account.Owner,
	}
}

func PresentSsoProvider(provider *api.SsoProvider) public.SsoProvider {
	return public.SsoProvider{
		Jwks:        provider.Jwks,
		BaseUrl:     provider.BaseUrl,
		TokenUrl:    provider.TokenUrl,
		ValidIssuer: provider.ValidIssuer,
	}
}
