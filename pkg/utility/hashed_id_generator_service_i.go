package utility

type HashedIdGeneratorServiceI interface {
	GenerateHashedId(id string, secret string) (string, error)
}
