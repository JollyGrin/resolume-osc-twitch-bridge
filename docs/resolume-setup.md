# Resolume Arena Setup Guide

This guide explains how to configure Resolume Arena to receive OSC commands from the Twitch-OSC bridge.

---

## 1. Enable OSC in Resolume

1. Open Resolume Arena
2. Go to **Arena → Preferences** (or `Cmd + ,` on Mac)
3. Click the **OSC** tab
4. Configure these settings:

| Setting | Value |
|---------|-------|
| OSC Input | **Enabled** |
| Input Port | `7000` |
| OSC Output | Optional (enable if you want to monitor) |
| Output Port | `7001` (if enabled) |

5. Click **OK** to save

---

## 2. Create a Text Alert Clip on Layer 1

### Step 2.1: Add a Text Source

1. In the **Sources** panel (bottom-left), go to the **Sources** tab
2. Scroll to find **Text Block** or **Text Animator** under the generators
   - **Text Block**: Simple static text (recommended for alerts)
   - **Text Animator**: Animated text effects
3. Drag **Text Block** to **Layer 1, Clip 1** (top-left clip slot)

### Step 2.2: Configure the Text Clip

1. Click on the clip you just created to select it
2. In the **Clip** panel (right side), you'll see the text parameters
3. Find the **Text** parameter - this is what we'll control via OSC
4. Set a default/placeholder text like `"Alert Text Here"`

### Step 2.3: Style the Text (Optional)

While the clip is selected, adjust:
- **Font**: Choose a readable font
- **Size**: Large enough to be visible on stream
- **Color**: Pick a color that stands out
- **Position**: Center or wherever you want alerts to appear

### Step 2.4: Set Clip Duration

1. Right-click the clip
2. Set **Transport** → **Timeline**
3. Set duration (e.g., 5 seconds for alerts)
4. Enable **Auto Pilot** if you want it to auto-stop after playing

---

## 3. Verify the OSC Address

The OSC address for your Text Block clip will be:

```
/composition/layers/1/clips/1/video/effects/textblock/effect/text/params/lines
```

To trigger (play) the clip:
```
/composition/layers/1/clips/1/connect
```

**Important**: OSC addresses vary by source type. Use Resolume's Shortcuts panel to find the exact address for your setup (see Troubleshooting).

### Finding OSC Addresses in Resolume

1. Go to **Arena → Preferences → OSC**
2. Enable **OSC Output** temporarily
3. Use an OSC monitor app (like **Protocol** on Mac, or **OSCDataMonitor**)
4. Interact with Resolume - the monitor shows the exact addresses

---

## 4. Test with OSC (Manual)

Before using our app, test OSC manually:

### Option A: Using `sendosc` (if installed)
```bash
# Trigger the clip
sendosc 127.0.0.1 7000 /composition/layers/1/clips/1/connect i 1

# Set text
sendosc 127.0.0.1 7000 /composition/layers/1/clips/1/video/source/textgenerator/params/text/value s "Test Alert!"
```

### Option B: Using our spike (see below)
Run the Go spike to test connectivity.

---

## 5. Layer Organization (Recommended)

For a clean setup, organize your layers:

| Layer | Purpose | Clips |
|-------|---------|-------|
| Layer 1 | Alerts | Follow, Subscribe, Raid, etc. |
| Layer 2 | Chat overlay | Chat messages (future) |
| Layer 3 | Video effects | Coin animations, etc. (future) |
| Layer 4+ | Your visuals | Your normal VJ content |

---

## Troubleshooting

### OSC Not Working?

1. **Check port**: Ensure Resolume is listening on port 7000
2. **Check firewall**: Allow UDP traffic on port 7000
3. **Check address**: OSC addresses are case-sensitive
4. **Check Resolume focus**: Some OSC commands require Resolume to be in focus

### Text Not Updating?

1. Ensure the clip has a **Text Block** or **Text Animator** source
2. The clip must exist at the specified layer/clip position
3. Try triggering the clip first, then setting text
4. **OSC address may be wrong** - use Shortcuts panel to find the correct address:
   - View → Shortcuts (or `Cmd+Shift+S`)
   - Select your text clip
   - Find the Text parameter and check its OSC address
   - Different sources have different paths (e.g., `textblock` vs `textgenerator`)

### Clip Not Playing?

1. Send value `1` (integer) to the `/connect` address
2. Check that the layer is not muted (solo mode)
3. Check clip transport settings (Timeline vs Loop)
