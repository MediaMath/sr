package sr

//Compatibility is the compatibility level from the schema registry.  See schema registry documentation for more details
type Compatibility string

const (
	//Zero is a zero value for compatibility and is not returned by the schema registry
	Zero = Compatibility("")

	//None means no compatibility between schemas is expected
	None = Compatibility("NONE")

	//Full means both Forward and Backward compatibility is expected
	Full = Compatibility("FULL")

	//Forward compatibility means all previous schemas can read the compatible schema
	Forward = Compatibility("FORWARD")

	//Backward compatibility means the compatible schema can read all previous schemas
	Backward = Compatibility("BACKWARD")
)
