package disk

func SameVolume(a, b string) (bool, error) {
	ka, err := volumeKey(existingDir(a))
	if err != nil {
		return false, err
	}
	kb, err := volumeKey(existingDir(b))
	if err != nil {
		return false, err
	}
	if ka == "" || kb == "" {
		return false, nil
	}
	return ka == kb, nil
}
