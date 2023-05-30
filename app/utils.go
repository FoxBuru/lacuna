package app

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aplr/lacuna/docker"
	"github.com/aplr/lacuna/pubsub"
)

func extractSubscriptions(container docker.Container) []pubsub.Subscription {
	subscriptions := make([]pubsub.Subscription, 0)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// Intermediate storage to hold subscriptions as we process labels
	subscriptionMap := make(map[string]*pubsub.Subscription)

	// Gather subscriptions by processing a container's labels
	for key, value := range container.Labels {
		keyParts := strings.Split(key, ".")

		// Check that the key starts with pubsub.subscription
		if len(keyParts) < 2 || keyParts[0] != labelPrefix || keyParts[1] != "subscription" {
			continue
		}

		// Check that the key has the correct number of parts
		if len(keyParts) != 4 {
			fmt.Printf("invalid subscription key: %s, must be in the format 'pubsub.subscription.<name>.<topic|endpoint>'\n", key)
			continue
		}

		// Check that the subscription name is valid
		if !nameRegex.MatchString(keyParts[2]) {
			fmt.Printf("invalid subscription name in key: %s, subscription name should be alphanumeric and may contain dashes\n", key)
			continue
		}

		name := strings.ToLower(keyParts[2])

		// Check if subscription already exists in the map
		if _, ok := subscriptionMap[name]; !ok {
			subscriptionMap[name] = &pubsub.Subscription{
				Service: container.Name(),
				Name:    name,
			}
		}

		// Assign the value to the correct field
		switch keyParts[3] {
		case "topic":
			subscriptionMap[name].Topic = value
		case "endpoint":
			subscriptionMap[name].Endpoint = value
		default:
			fmt.Printf("skipping invalid subscription key: %s, must be one of 'topic' or 'endpoint'\n", key)
		}

	}

	// Convert map to slice, only consider valid subscriptions
	for _, subscription := range subscriptionMap {
		if subscription.Topic != "" && subscription.Endpoint != "" {
			// Only include subscriptions with both topic and endpoint populated
			subscriptions = append(subscriptions, *subscription)
		} else {
			fmt.Printf("skipping incomplete subscription: %s, both topic and endpoint must be provided\n", subscription.Name)
		}
	}

	return subscriptions
}
