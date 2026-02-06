# Learnings

Lessons learned during development of the Resolume Twitch OSC bridge.

---

## OSC Address Discovery (2024)

**Problem**: OSC text address didn't work - Resolume received messages but text didn't change.

**Root Cause**: The OSC address path varies depending on how the text source is added to Resolume:

| Source Type | OSC Address Pattern |
|-------------|---------------------|
| Text Block (as effect) | `/composition/layers/{L}/clips/{C}/video/effects/textblock/effect/text/params/lines` |
| Text Generator (as source) | `/composition/layers/{L}/clips/{C}/video/source/textgenerator/params/text/value` |

Most online examples use `textgenerator` but dragging "Text Block" from Sources creates an effect-based setup.

**Solution**: Always verify OSC addresses using Resolume's Shortcuts panel:
1. View → Shortcuts (`Cmd+Shift+S`)
2. Select the clip/parameter
3. Check the actual OSC address shown

**Lesson**: Don't assume OSC addresses from documentation or examples. Resolume's internal structure varies by how clips are created. Always verify with the Shortcuts panel.

---

## OSC Debugging Tips

1. **Enable OSC Input monitor** in Resolume Preferences → OSC to confirm messages are arriving
2. **Use Shortcuts panel** to find exact OSC addresses (View → Shortcuts)
3. **Enable OSC Output** and use an external monitor (like Protocol app) to see what addresses Resolume uses when you interact with it manually
4. **Trigger works but text doesn't?** Address path is wrong - check if it's `effects/` vs `source/`

---

## Resolume Text Block Setup

When creating a text overlay clip:
1. Drag "Text Block" from Sources → Generators to a clip slot
2. This creates it as an **effect**, not a source
3. The OSC path will include `video/effects/textblock/effect/text/params/lines`
4. The clip trigger address remains `/composition/layers/{L}/clips/{C}/connect`
