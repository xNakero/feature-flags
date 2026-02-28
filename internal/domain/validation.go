package domain

import "fmt"

func ValidateFlagName(name string) error {
	if err := validateNotEmpty(name); err != nil {
		return err
	}
	if err := validateMaxLength(name); err != nil {
		return err
	}
	if err := validateStartsWithLetter(name); err != nil {
		return err
	}
	if err := validateNoLeadingTrailingHyphen(name); err != nil {
		return err
	}
	return validateAllowedChars(name)
}

func ValidateFlagValue(flagType FlagType, flagValue FlagValue) error {
	switch flagType {
	case FlagTypeBoolean:
		return validateBooleanValue(flagValue)
	case FlagTypeNumeric:
		return validateNumericValue(flagValue)
	}
	return nil
}

func validateNotEmpty(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name must not be empty: %w", ErrInvalidName)
	}
	return nil
}

func validateMaxLength(name string) error {
	if len(name) > 63 {
		return fmt.Errorf("name must not exceed 63 characters: %w", ErrInvalidName)
	}
	return nil
}

func validateStartsWithLetter(name string) error {
	if name[0] < 'a' || name[0] > 'z' {
		return fmt.Errorf("name must start with a lowercase letter: %w", ErrInvalidName)
	}
	return nil
}

func validateNoLeadingTrailingHyphen(name string) error {
	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("name must not start or end with a hyphen: %w", ErrInvalidName)
	}
	return nil
}

func validateAllowedChars(name string) error {
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return fmt.Errorf("name must contain only lowercase letters, digits, and hyphens: %w", ErrInvalidName)
		}
	}
	return nil
}

func validateBooleanValue(flagValue FlagValue) error {
	if flagValue.Bool == nil {
		return fmt.Errorf("boolean flag requires a bool value: %w", ErrTypeMismatch)
	}
	return nil
}

func validateNumericValue(flagValue FlagValue) error {
	if flagValue.Numeric == nil {
		return fmt.Errorf("numeric flag requires a numeric value: %w", ErrTypeMismatch)
	}
	return nil
}
