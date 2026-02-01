package model

import "time"

type DeliveryPolicy interface {
	Deadline(now time.Time) time.Time
}

type fixedDurationPolicy struct {
	duration time.Duration
}

func (p fixedDurationPolicy) Deadline(now time.Time) time.Time {
	return now.Add(p.duration)
}

type DeliveryTimeFactory struct {
	defaultPolicy DeliveryPolicy
	policies      map[TransportType]DeliveryPolicy
}

func NewDeliveryTimeFactory(onFoot, scooter, car time.Duration) *DeliveryTimeFactory {
	defaultPolicy := fixedDurationPolicy{duration: onFoot}
	return &DeliveryTimeFactory{
		defaultPolicy: defaultPolicy,
		policies: map[TransportType]DeliveryPolicy{
			TransportOnFoot:  fixedDurationPolicy{duration: onFoot},
			TransportScooter: fixedDurationPolicy{duration: scooter},
			TransportCar:     fixedDurationPolicy{duration: car},
		},
	}
}

func (f *DeliveryTimeFactory) ForTransport(transport TransportType) DeliveryPolicy {
	if policy, ok := f.policies[transport]; ok {
		return policy
	}
	return f.defaultPolicy
}
