package util

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ValidateContactAddress(typeKey, countryCodeKey string) validator.String {
	return &contactAddressValidator{stringDescriptor{"TODO"}, typeKey, countryCodeKey}
}

type contactAddressValidator struct {
	stringDescriptor
	typeKey, countryCodeKey string
}

func (v contactAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var typeConfig types.String
	d := req.Config.GetAttribute(ctx, path.Root(v.typeKey), &typeConfig)
	resp.Diagnostics.Append(d...)

	var countryCode types.Int64
	d = req.Config.GetAttribute(ctx, path.Root(v.countryCodeKey), &countryCode)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}
	t := typeConfig.ValueString()
	code := int(countryCode.ValueInt64())
	addr := req.ConfigValue.ValueString()

	if t == "sms_contact_method" || t == "phone_contact_method" {
		// Validation logic based on https://support.pagerduty.com/docs/user-profile#phone-number-formatting
		maxLength := 40

		if len(addr) > maxLength {
			resp.Diagnostics.AddError("phone numbers may not exceed 40 characters", addr)
			return
		}

		if !phoneOnlyAllowedChars.MatchString(addr) {
			resp.Diagnostics.AddError(
				"phone numbers may only include digits from 0-9 and the symbols: comma (,), asterisk (*), and pound (#)",
				addr,
			)
			return
		}

		isMexicoNumber := code == 52
		if t == "sms_contact_method" && isMexicoNumber && strings.HasPrefix(addr, "1") {
			resp.Diagnostics.AddError(
				"Mexico-based SMS numbers should be free of area code prefixes",
				fmt.Sprintf("Please remove the leading 1 in the number %q", addr),
			)
			return
		}

		isTrunkPrefixNotSupported := map[int]string{
			33: "0", // France (33-0)
			40: "0", // Romania (40-0)
			44: "0", // UK (44-0)
			45: "0", // Denmark (45-0)
			49: "0", // Germany (49-0)
			61: "0", // Australia (61-0)
			66: "0", // Thailand (66-0)
			91: "0", // India (91-0)
			1:  "1", // North America (1-1)
		}

		prefix, ok := isTrunkPrefixNotSupported[code]
		if ok && strings.HasPrefix(addr, prefix) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Trunk prefixes are not supported for following countries and regions: France, Romania, UK, Denmark, Germany, Australia, Thailand, India and North America, so must be formatted for international use without the leading %s", prefix),
				"",
			)
			return
		}
	}
}

var phoneOnlyAllowedChars = regexp.MustCompile(`^[0-9,*#]+$`)
