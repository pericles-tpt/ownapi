import { WordEnd, Trie, aLessThanOnCharIdxLookup, charIdxLookup, getDefaultTrie } from "./structure";

// TableFieldTrie, uses either a recursive or iterative search for different values of n, prioritising accuracy (i.e. closest to target term) of results
// NOTE: When the recursive algorithm is chosen it sets the `retrieveNClosestResults` limit to 2_000_000_000 (basically all results). This is because 
// RECURSIVE doesn't retrieve results in "closest match" order (unlike ITERATIVE), probably due to its DFS traversal, and needs to:
//       1. Retrieve all matches
//       2. Sort matches in "closest match to term" order
//       3. Return a slice of the first `retrieveNClosestResults` that the user requested
export class TableFieldTrie {
    private root: Trie;
    
    constructor(words: string[]) {
        this.regenerate(words);
    }

    public regenerate(words: string[]) {
        this.root = this.generateTrieForWords(words);
    }

    private generateTrieForWords(words: string[]): Trie {
        const ret: Trie = getDefaultTrie();
        words.forEach((w, i) => {
            this.walkGenerateTrieForWord(ret, i, w.toLowerCase());
        });
        return ret;
    }

    private walkGenerateTrieForWord(root: Trie, arrIdx: number, lowerWord: string) {
        if (lowerWord.length === 0) {
            root.wordEnd = {
                wordEnd: true,
                arrIndex: arrIdx
            }
            return;
        }

        const firstCharIdx = charIdxLookup.get(lowerWord[0]);
        if (firstCharIdx === undefined) {
            console.warn(`failed to find char '${lowerWord[0]}' in charIdxLookup`)
            return;
        }

        if (root.isPopulatedAt[firstCharIdx] === -1) {
            root.nodes.push([lowerWord[0], getDefaultTrie()]);
            
            // Sort on every push to make sure nodes maintain order w.r.t mapping
            root.nodes.sort((a, b) => {
                return (a[0].length < b[0].length || (a[0].length === b[0].length && aLessThanOnCharIdxLookup(a[0], b[0]))) ? -1 : 1;
            })
            root.nodes.forEach((n, i) => {
                root.isPopulatedAt[charIdxLookup.get(n[0]) as number] = i;
            });
        }
        this.walkGenerateTrieForWord(root.nodes[root.isPopulatedAt[firstCharIdx]][1], arrIdx, lowerWord.slice(1,));
    }


    // NOTE: Recursive seems to be faster when the KNOWN number of results is > 50, BUT this hasn't been thoroughly tested
    public findNClosestWordsInTrie(term: string, retrieveNClosestResults: number, forceRecursive: boolean = false, forceIterative: boolean = false): [string, WordEnd][] {
        term                               = term.toLowerCase();
        const allCharsValid                = term.split("").reduce((b, ch) => b && (charIdxLookup.get(ch) !== undefined), true)
        let res: [string, WordEnd][] = [];
        if (term.length === 0 || retrieveNClosestResults <= 0 || !allCharsValid) {
            return res;
        }

        if (!forceIterative && (forceRecursive || retrieveNClosestResults > 50)) {
            res = this.findNClosestWordsInTrieRec(this.root, term, "", 2_000_000_000, [])[1].sort((a, b) => {
                return (a[0].length < b[0].length || (a[0].length === b[0].length && aLessThanOnCharIdxLookup(a[0], b[0]))) ? -1 : 1;
            }).slice(0, retrieveNClosestResults);
        } else {
            res = this.findNClosestWordsInTrieIter(term, retrieveNClosestResults);
        }
    
        return res;
    }

    private findNClosestWordsInTrieIter(term: string, upToNClosest: number): [string, WordEnd][] {
        const ret: [string, WordEnd][] = [];
        const validNodeQ: Trie[]             = [this.root];
        const validWordQ: string[]           = [""];

        while (validNodeQ.length > 0 && upToNClosest > 0) {
            const currNode = validNodeQ.shift() as Trie;
            
            // <  0 -> at end of word, so look through ALL nodes to find closest matches
            // >= 0 -> more letter left in word, go directly to that index in next node
            const isCharsLeft = term.length > 0;
            let firstCharIdx  = -1
            if (isCharsLeft) {
                firstCharIdx = charIdxLookup.get(term[0]) as number;
                validWordQ[0] += term[0];
                term = term.slice(1,);
                
                // Next char in word isn't in tree
                if (currNode.isPopulatedAt[firstCharIdx] === -1) {
                    return ret;
                }

                validNodeQ.push(currNode.nodes[currNode.isPopulatedAt[firstCharIdx]][1]);
                continue;
            }
            
            // Check if this node is a word ending
            const currWord = validWordQ.shift() as string;
            if (currNode.wordEnd) {
                ret.push([currWord, currNode.wordEnd]);
                upToNClosest--
            }
            
            for (let i = 0; i < currNode.nodes.length; i++) {
                const n = currNode.nodes[i];
                validWordQ.push(currWord + n[0]);
                validNodeQ.push(n[1]);
            }
        }

        return ret;
    }

    private findNClosestWordsInTrieRec(root: Trie, wordDec: string, wordInc: string, upToNClosest: number, results: [string, WordEnd][]): [number, [string, WordEnd][]] {
        if (upToNClosest === 0) {
            return [upToNClosest, results];
        }
        
        const isCharsLeft = wordDec.length > 0;
        if (isCharsLeft) {
            let firstCharIdx = charIdxLookup.get(wordDec[0]) as number;

            // Next char in word isn't in tree
            const nextNodeIdx = root.isPopulatedAt[firstCharIdx]
            if (nextNodeIdx === -1) {
                return [upToNClosest, results];
            }

            wordInc = wordInc + wordDec[0];
            wordDec = wordDec.slice(1,)
            return this.findNClosestWordsInTrieRec(root.nodes[nextNodeIdx][1], wordDec, wordInc, upToNClosest, results);
        }
        
        // Check if this node is a word ending
        if (root.wordEnd.wordEnd) {
            results.push([wordInc, root.wordEnd]);
            upToNClosest--
        }
        
        for (let i = 0; i < root.nodes.length; i++) {
            [upToNClosest, ] = this.findNClosestWordsInTrieRec(root.nodes[i][1], wordDec, wordInc + root.nodes[i][0], upToNClosest, results);
        }
        return [upToNClosest, results];
    }
}