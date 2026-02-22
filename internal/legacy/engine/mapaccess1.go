package engine

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/abiosoft/ishell/v2"
)

func (m *Meta[K, V]) registerMapAccess1() {
	m.Console.RegisterCommand(
		"mapaccess1",
		"mapaccess1 <key> — simulate mapaccess1 step by step",
		func(ctx *ishell.Context) {
			args := ctx.Args
			if len(args) < 1 {
				ctx.PrintlnLogWarn("Usage: mapaccess1 <key>")
				return
			}

			key, err := ParseValue[K](args[0])
			if err != nil {
				ctx.PrintlnLogError("Invalid key:", err)
				return
			}

			h := GetHmap(m.Map)
			if h == nil || h.count == 0 {
				ctx.PrintlnLogWarn("map is nil or empty → mapaccess1 returns zero value immediately")
				return
			}

			if mapType == nil {
				addr := GetMapType(m.Map)
				mapType = (*maptype)(unsafe.Pointer(addr))
			}

			a := func(code, s string) string { return "\x1b[" + code + "m" + s + "\x1b[0m" }
			dim := func(s string) string { return a("2", s) }
			cyan := func(s string) string { return a("96", s) }
			yellow := func(s string) string { return a("93", s) }
			green := func(s string) string { return a("92", s) }
			red := func(s string) string { return a("91", s) }
			bold := func(s string) string { return a("1", s) }
			orange := func(s string) string { return a("38;2;255;165;0", s) }

			stepN := 0
			header := func(label string) {
				stepN++
				ctx.Printf("\n%s\n", cyan(bold(fmt.Sprintf("📎 step %d · %s", stepN, label))))
			}
			yes := func(s string) { ctx.Printf("  %s\n", green("✓ "+s)) }
			no := func(s string) { ctx.Printf("  %s\n", red("✗ "+s)) }
			note := func(s string) { ctx.Printf("  %s\n", dim("· "+s)) }

			toBin := func(v uint64, bits int) string {
				return fmt.Sprintf("%0*b", bits, v)
			}

			thLabel := func(th uint8) string {
				switch th {
				case 0:
					return "emptyRest"
				case 1:
					return "emptyOne"
				case 2:
					return "evacX"
				case 3:
					return "evacY"
				case 4:
					return "evacEmpty"
				default:
					return fmt.Sprintf("%02x", th)
				}
			}

			// ── step 1 hash seed ────────────────────────────────────────
			header("hash seed")
			ctx.Printf("  hmap.hash0 = %s\n", yellow(fmt.Sprintf("%d", h.Hash0)))
			ctx.Printf("  hmap.hash0 = %s\n", yellow(fmt.Sprintf("0b%s", toBin(uint64(h.Hash0), 32))))

			// ── step 2 compute hash ─────────────────────────────────────
			hash := FullHash(m, key)
			header("compute hash")
			ctx.Printf("  HashFunc(%s, %s)\n",
				orange(fmt.Sprintf("%v", key)),
				yellow(fmt.Sprintf("%d", h.Hash0)))
			ctx.Printf("       = %s\n", yellow(fmt.Sprintf("%d (dec)", hash)))
			ctx.Printf("       = %s\n", yellow(fmt.Sprintf("%s (bin)", toBin(uint64(hash), 64))))

			// ── step 3 bucket mask ──────────────────────────────────────
			B := h.B
			mask := uintptr(1)<<B - 1
			header("bucket mask")
			ctx.Printf("  h.B = %s\n", yellow(fmt.Sprintf("%d  →  2^B = %d buckets", B, uintptr(1)<<B)))
			ctx.Printf("  mask = (1<<%d) - 1\n", B)
			ctx.Printf("       = %s\n", yellow(fmt.Sprintf("%d", mask)))
			ctx.Printf("       = %s\n", yellow(fmt.Sprintf("0b%s", toBin(uint64(mask), int(B)))))

			// ── step 4 select bucket ────────────────────────────────────
			bucketIdx := hash & mask
			header("select bucket")

			const pw, bw = 9, 64
			hashBin := toBin(uint64(hash), bw)
			maskBin := toBin(uint64(mask), bw)
			resBin := toBin(uint64(bucketIdx), bw)
			sep := dim(strings.Repeat("─", pw+bw))

			ctx.Printf("  hash : %s\n", yellow(hashBin))
			ctx.Printf("& mask : %s\n", yellow(maskBin))
			ctx.Printf("%s\n", sep)
			ctx.Printf("       = %s\n", yellow(resBin))
			ctx.Printf("  bucket index = %s\n", bold(orange(fmt.Sprintf("%d", bucketIdx))))

			b := (*_bucket_[K, V])(unsafe.Add(h.buckets, bucketIdx*m.bucketSizeof))

			// ── step 5 growing check ────────────────────────────────────
			header("growing check  (h.oldbuckets)")
			if h.oldbuckets == nil {
				note("oldbuckets = nil → no grow in progress, use b as-is")
			} else {
				ctx.Printf("  %s\n", yellow("oldbuckets != nil → map is growing!"))
				sameSizeGrow := h.flags&8 != 0
				var oldMask uintptr
				if sameSizeGrow {
					note("flags & sameSizeGrow=1 → same-size grow → oldmask = mask")
					oldMask = mask
				} else {
					oldMask = mask >> 1
					note(fmt.Sprintf("regular grow → oldmask = mask>>1 = %d", oldMask))
				}
				oldIdx := hash & oldMask
				oldb := (*_bucket_[K, V])(unsafe.Add(h.oldbuckets, oldIdx*m.bucketSizeof))
				ctx.Printf("  old bucket idx = %s\n", yellow(fmt.Sprintf("%d", oldIdx)))
				th0 := oldb.tophash[0]
				ctx.Printf("  oldb.tophash[0] = %s\n", yellow(thLabel(th0)))
				// evacuatedX=2, evacuatedY=3, evacuatedEmpty=4
				if th0 == 2 || th0 == 3 || th0 == 4 {
					yes(fmt.Sprintf("evacuated → use new b = buckets[%d]", bucketIdx))
				} else {
					no(fmt.Sprintf("not yet evacuated → use old b = oldbuckets[%d]", oldIdx))
					b = oldb
				}
			}

			// ── step 6 compute tophash ──────────────────────────────────
			top := tophash(hash)
			ptrBits := 8 * int(unsafe.Sizeof(uintptr(0)))
			header(fmt.Sprintf("compute tophash  (hash >> %d)", ptrBits-8))
			ctx.Printf("  top = uint8(hash >> %d)\n", ptrBits-8)
			raw := uint8(hash >> uint(ptrBits-8))
			if raw < 5 {
				note(fmt.Sprintf("raw=0x%02x < minTopHash(5) → top = raw+5 = 0x%02x", raw, top))
			}
			ctx.Printf("  top = %s\n", yellow(fmt.Sprintf("0x%02x  (%d)", top, top)))

			// ── step 7 bucket chain scan ────────────────────────────────
			header("bucket chain scan")
			found := false
			chainIdx := 0

		CHAIN:
			for curr := b; curr != nil; curr = (*_bucket_[K, V])(curr.overflow) {
				ctx.Printf("\n  ╭─ ")
				ctx.Printf("%s\n", cyan(fmt.Sprintf("chain node #%d", chainIdx)))
				ctx.Printf("  │  tophash = [")
				for i := 0; i < 8; i++ {
					ctx.Printf("%s", yellow(thLabel(curr.tophash[i])))
					if i < 7 {
						ctx.Printf(" ")
					}
				}
				ctx.Printf("]\n  │\n")

				for i := 0; i < 8; i++ {
					th := curr.tophash[i]
					ctx.Printf("  │  slot[%d]", i)

					switch {
					case th == 0: // emptyRest
						no("emptyRest → no more entries, break outer loop")
						break CHAIN

					case th == 1: // emptyOne
						note("emptyOne → slot empty, skip")
						continue

					case th != top:
						no(fmt.Sprintf("%s ≠ %s → next", thLabel(th), thLabel(top)))
						continue
					}

					// th == top: tophash match
					yes(fmt.Sprintf("%02x = %02x → hash match, check key equality", th, top))

					k := curr.keys[i]
					ctx.Printf("  │  %s =? %s",
						orange(fmt.Sprintf("%v", k)),
						orange(fmt.Sprintf("%v", key)))
					if any(k) == any(key) {
						yes("key match")
						ctx.Printf("  │  value = %s\n", yellow(fmt.Sprintf("%v", curr.values[i])))
						found = true
						break CHAIN
					}
					no("key mismatch (hash collision) → continue")
				}

				ctx.Printf("  │\n  ╰─ overflow → ")
				if curr.overflow != nil {
					ctx.Printf("%s\n", yellow(fmt.Sprintf("0x%x", uintptr(curr.overflow))))
				} else {
					ctx.Printf("%s\n", dim("nil (end of chain)"))
				}
				chainIdx++
			}

			// ── result ────────────────────────────────────────────────────
			header("result")
			if found {
				yes(bold("FOUND"))
			} else {
				no("not found → return &zeroVal[0]  (zero value)")
			}
		},
	)
}
