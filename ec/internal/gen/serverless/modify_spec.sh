#!/bin/env bash

# Step 1: Apply plan modifiers for string attributes
jq --slurpfile plan_modifiers string_use_state_for_unknown.json 'def applies_to($item): $plan_modifiers[0].applies_to[] | any(. == $item; .); (.resources[] | .schema.attributes[] | select(applies_to(.name))).string.plan_modifiers |= $plan_modifiers[0].add' spec.json > /tmp/with-strings.json

# Step 2: Convert traffic_filters from list_nested (with nested id object) to set of strings
# This makes the schema simpler - users just provide a list of traffic filter IDs
jq '(.resources[].schema.attributes[] | select(.name == "traffic_filters")) |= {
  "name": "traffic_filters",
  "set": {
    "computed_optional_required": "optional",
    "description": "Traffic filters to associate with this project. Traffic filters are IDs of traffic filter resources.",
    "element_type": {
      "string": {}
    }
  }
}' /tmp/with-strings.json > ./spec-mod.json
