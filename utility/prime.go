package utility

import (
	"fmt"
	"math"
)

var (
	primes []int
)

func InitPrimes(limit int) {
	primes = primeSieve(limit)
}

func primeSieve(limit int) []int {
	if limit < 2 {
		return []int{}
	}

	var (
		p = 2
		// Not memory efficient, but fine for my calculations up to a small limit
		composites = make([]bool, limit+1)
		primes     = make([]int, 0, limit/10)
	)

	// Handle 2 first, increment to 3, then following iterations with inc by 2 (to skip evens)
	primes = append(primes, 2)
	flagComposites(p, limit, composites)
	p++

	for p <= limit {
		isPrime := !composites[p]
		if isPrime {
			primes = append(primes, p)
			flagComposites(p, limit, composites)
		}
		p += 2
	}

	return primes
}

func flagComposites(currPrime int, limit int, composites []bool) {
	var (
		i    = currPrime
		curr = i * currPrime
	)
	for curr < limit {
		composites[curr] = true
		i++
		curr = i * currPrime
	}
}

// GCD, calculates the greatest common denominator of a set of integers, primes are
// precomputed with InitPrimes. Providing a number n: sqrt(n) > primes[len(primes)-1]
// will throw an error
//
// A divideByUpTo > 1 will iterate the computed factors and multiply them until the
// the threshold is reached, the returned gcd will be divided by the result and the
// multiplied factors will be removed from the start of the returned array
func GCD(set []int, divideByUpTo int) (int, []int, error) {
	setCopy := make([]int, len(set))
	copy(setCopy, set)

	max := -1
	for _, num := range setCopy {
		if num > max {
			max = num
		}
	}

	var (
		// Source (kind of): https://www.cambridge.org/core/journals/canadian-mathematical-bulletin/article/abs/number-of-prime-factors-with-a-given-multiplicity/5CDE211A9FA6257416BEFB71F6248D73
		// ^ the formula mentioned above is correct for calculating UNIQUE prime factors, not with multiplicity, just need a single log(n) to estimate factors with multiplicity
		estNumFactors = int(math.Log(float64(max)) + 1)
		factors       = make([]int, 0, estNumFactors)
		gcd           = 1
	)
	// Check that provided `primes` are sufficient for up to `max`
	var (
		maxPrimeLimit    = int(math.Sqrt(float64(max)))
		maxComputedPrime = primes[len(primes)-1]
	)
	if maxPrimeLimit > maxComputedPrime {
		return gcd, factors, fmt.Errorf("invalid `max` (%d) provided in set, it's maximum prime is greater than the max precomputed prime: %d > %d", max, maxPrimeLimit, maxComputedPrime)
	}
	// Only need to traverse primes up to primeLimit
	stopBeforeIdx := len(primes)
	for i, p := range primes {
		if p > maxPrimeLimit {
			stopBeforeIdx = i
			break
		}
	}

	for i := 0; i < stopBeforeIdx; i++ {
		p := primes[i]
		pFactorsAll := true
		for _, num := range setCopy {
			pFactorsAll = pFactorsAll && (num%p == 0)
		}

		if pFactorsAll {
			factors = append(factors, p)
			gcd *= p
			for i := range setCopy {
				setCopy[i] /= p
			}
			i--
		}
	}

	// After calculating gcd, if a valid divideByUpTo is provided then try to remove factors and divide gcd up to that number
	var (
		divideBy = 1
		i        = 0
	)
	for divideBy < divideByUpTo && i < len(factors) {
		m := divideBy * factors[i]
		if m > divideByUpTo {
			break
		}
		divideBy = m
		i++
	}
	return gcd / divideBy, factors[i:], nil
}
