package schema

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// ignoreCaseDiffSuppressFunc is a DiffSuppressFunc from helper/schema that is
// used to ignore any case-changes in a return value.
func IgnoreCaseDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	log.Printf("[INFO] IgnoreCaseDiffSuppressFunc: (%q, %q, %q)", k, old, new)
	log.Printf("[INFO] k = % x", k)
	log.Printf("[INFO] old = % x", old)
	log.Printf("[INFO] new = % x", new)
	log.Printf("[INFO] strings.ToLower(te%q [old]) = %q", old, strings.ToLower(old))
	log.Printf("[INFO] strings.ToLower(%q [new]) = %q", new, strings.ToLower(new))
	log.Printf("[INFO] strings.ToLower(old) == strings.ToLower(new)? %v", strings.ToLower(old) == strings.ToLower(new))
	return strings.ToLower(old) == strings.ToLower(new)
}

// ignoreCaseStateFunc is a StateFunc from helper/schema that converts the
// supplied value to lower before saving to state for consistency.
func IgnoreCaseStateFunc(val interface{}) string {
	return strings.ToLower(val.(string))
}
