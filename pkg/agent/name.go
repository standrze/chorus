package agent

import (
	"fmt"
	"math/rand"
)

func GenerateAgentName() string {
	adjectives := []string{
		"admiring", "adoring", "agitated", "amazing", "angry", "awesome",
		"beautiful", "blissful", "bold", "boring", "brave", "busy",
		"charming", "clever", "cool", "compassionate", "competent",
		"condescending", "confident", "cranky", "crazy", "dazzling",
		"determined", "distracted", "dreamy", "eager", "ecstatic",
		"elastic", "elated", "elegant", "eloquent", "epic", "exciting",
		"fervent", "festive", "flamboyant", "focused", "friendly",
		"frosty", "funny", "gallant", "gifted", "goofy", "gracious",
		"great", "happy", "hardcore", "heuristic", "hopeful",
		"hungry", "infallible", "inspiring", "intelligent", "interesting",
		"jolly", "jovial", "keen", "kind", "laughing", "loving",
		"lucid", "magical", "mystifying", "modest", "musing", "naughty",
		"nervous", "nice", "nifty", "nostalgic", "objective", "optimistic",
		"peaceful", "pedantic", "pensive", "practical", "priceless",
		"quirky", "quizzical", "recursing", "relaxed", "reverent",
		"romantic", "sad", "serene", "sharp", "silly", "sleepy",
		"stoic", "strange", "stupefied", "suspicious", "sweet",
		"tender", "thirsty", "trusting", "unruffled", "upbeat",
		"vibrant", "vigilant", "vigorous", "wizardly", "wonderful",
		"xenodochial", "youthful", "zealous", "zen",
	}

	nouns := []string{
		"albattani", "allen", "almeida", "antonelli", "agnesi", "archimedes",
		"ardinghelli", "aryabhata", "austin", "babbage", "banach", "banzai",
		"bardeen", "bartik", "bassi", "beaver", "bell", "benz", "bhabha",
		"bhaskara", "black", "blackwell", "bohr", "booth", "borg", "bose",
		"boyd", "brahmagupta", "brattain", "brown", "buck", "burnell",
		"cannon", "carson", "cartwright", "carver", "cerf", "chandrasekhar",
		"chaplygin", "chatelet", "chatterjee", "chebyshev", "clarke", "cohen",
		"colden", "cori", "cray", "curie", "darwin", "davinci", "dewdney",
		"dhawan", "diffie", "dijkstra", "dirac", "driscoll", "dubinsky",
		"easley", "edison", "einstein", "elbakyan", "elgamal", "elion",
		"ellis", "engelbart", "euclid", "euler", "faraday", "feistel",
		"fermat", "fermi", "feynman", "franklin", "gagarin", "galileo",
		"galois", "ganguly", "gates", "gauss", "germain", "goldberg",
		"goldstine", "goldwasser", "golick", "goodall", "gould", "greider",
		"grothendieck", "haibt", "hamilton", "haslett", "hawking", "hellman",
		"heisenberg", "hermann", "herschel", "hertz", "heyrovsky", "hodgkin",
		"hofstadter", "hoover", "hopper", "hugle", "hypatia", "ishizaka",
		"jackson", "jang", "jemison", "jennings", "jepsen", "johnson",
		"joliot", "jones", "kalam", "kare", "keldysh", "keller", "kepler",
		"khayyam", "khorana", "kilby", "kirch", "knuth", "kowalevski",
		"lalande", "lamarr", "lamport", "leakey", "leavitt", "lederberg",
		"lehmann", "lewin", "lichterman", "liskov", "lovelace", "lumiere",
		"mahavira", "margulis", "matsumoto", "maxwell", "mayer", "mccarthy",
		"mcclintock", "mclaren", "mclean", "mcnulty", "meitner", "mendel",
		"mendeleev", "meninsky", "merkle", "mestorf", "mirzakhani", "montalcini",
		"moore", "morse", "murdock", "napier", "nash", "neumann", "newton",
		"nightingale", "nobel", "noether", "northcutt", "noyce", "panini",
		"pare", "pascal", "pasteur", "payne", "perlman", "pike", "poincare",
		"poitras", "proskuriakova", "ptolemy", "raman", "ramanujan", "ride",
		"ritchie", "rhodes", "robinson", "roentgen", "rosalind", "rubin",
		"russell", "saha", "sammet", "sanderson", "satoshi", "shamir",
		"shannon", "shaw", "shirley", "shockley", "shtern", "sinoussi",
		"snyder", "spence", "stonebraker", "sutherland", "swanson", "swartz",
		"swirles", "taussig", "tereshkova", "tesla", "tharp", "thompson",
		"torvalds", "tu", "turing", "varahamihira", "vaughan", "visvesvaraya",
		"volhard", "villani", "wescoff", "wilbur", "wiles", "williams",
		"williamson", "wilson", "wing", "wozniak", "wright", "wu", "yalow",
		"yonath", "zhukovsky",
	}

	adjective := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	return fmt.Sprintf("%s_%s", adjective, noun)
}
