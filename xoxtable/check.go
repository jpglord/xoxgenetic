package xoxtable

func IsGameFinish(table Table) bool {
	return table.IsFull()
}

func IsGameWin(t Table) bool {
	if t[0] == t[1] && t[1] == t[2] && t[0] != E {
		return true
	}

	if t[0] == t[3] && t[3] == t[6] && t[0] != E {
		return true
	}

	if t[0] == t[4] && t[4] == t[8] && t[0] != E {
		return true
	}

	if t[1] == t[4] && t[4] == t[7] && t[1] != E {
		return true
	}

	if t[2] == t[5] && t[5] == t[8] && t[2] != E {
		return true
	}

	if t[2] == t[4] && t[4] == t[6] && t[2] != E {
		return true
	}

	if t[3] == t[4] && t[4] == t[5] && t[3] != E {
		return true
	}

	if t[6] == t[7] && t[7] == t[8] && t[6] != E {
		return true
	}

	return false
}
