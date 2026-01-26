# The Expedition

> *A narrative journey through Chimborazo's architecture, told in the voice of Alexander von Humboldt upon his return to Paris, 1805.*

---

## Prologue: The View from 19,286 Feet

*Paris, Winter 1805*

I have been back in Europe for scarcely a year, yet already the memories of the journey threaten to become mere anecdotes—tales for salon conversation, stripped of their meaning. I must write while the altitude sickness still echoes in my lungs, while the Orinoco's insects still itch in memory, while Chimborazo's silhouette remains sharp against the inside of my eyelids.

I climbed that mountain in June of 1802. At 19,286 feet—higher than any European had ever stood—an impassable crevasse blocked the final thousand feet to the summit. We could see our goal but could not reach it.

And yet.

From that height, I beheld something more valuable than a summit: *connection*. The vegetation zones arranged themselves in bands below me, each climate creating its own world of life. Temperature, pressure, humidity, and biology formed a single system. I sketched what I saw, and that sketch—the Naturgemälde—became the foundation of everything I now understand about the unity of nature.

This toolkit bears the mountain's name because it pursues the same goal: to transform raw data into revelations. We may never reach the perfect map, but the view from where we stand is worth the climb.

---

## I. Preparation

*Madrid, March 1799*

Before one sets foot on foreign soil, one must prepare. I spent months in Madrid securing what I needed: instruments from the finest makers in Europe—chronometers, sextants, barometers, thermometers, cyanometers for measuring the blueness of the sky. More importantly, I secured *permission*: a passport from King Charles IV granting me unprecedented access to every corner of Spanish America.

Without preparation, the expedition would have failed before it began. A broken chronometer means useless longitude readings. A missing permit means arrest at the first colonial checkpoint.

### The Recipe as Expedition Plan

In Chimborazo, the recipe serves as your expedition plan. It declares:
- What data you seek (sources)
- What transformations you require (operations)
- What you wish to render (output)
- The bounds of your investigation

```yaml
name: Vermont Waters
description: Lakes and boundaries of the Green Mountain State
version: 1

sources:
  state:
    uri: "census://state/VT"
  counties:
    uri: "census://county/VT"
  water:
    uri: "census://water/VT"

output:
  bounds: [-73.5, 42.7, -71.5, 45.1]
  formats:
    - type: svg
      page_size: "12x18"
```

Before a single HTTP request is made, the recipe is validated. Are all sources reachable? Are the operations recognized? Do the layer references resolve? These checks happen in **Preparation**, before Acquisition begins.

### What Smith Taught Me

William Smith, the English geologist who mapped his nation's strata in 1815, understood something I only grasped later: *order matters*. His breakthrough was recognizing that rock layers always appear in the same sequence—that coal lies above limestone, which lies above older stone.

In cartography, layer order matters equally. Place your regional backgrounds at the bottom, your water bodies above, your boundaries next, your roads at the very top. Each stratum serves its purpose; none can be omitted without loss.

The recipe establishes this order. The `order` field on each layer determines what renders on top of what. Get the order wrong, and your water disappears beneath your land.

---

## II. Acquisition

*Cumaná, Venezuela, July 1799*

We landed at Cumaná on July 16, 1799—my first steps on South American soil. Within hours, I had collected dozens of specimens. Within days, hundreds. The abundance was overwhelming.

But abundance is not knowledge. I learned quickly that not all sources are equal. The local tales of "bottomless" lakes proved false when I sounded them myself. The colonial maps, drawn by officials who had never left their offices, were riddled with errors. To know a thing truly, I had to measure it myself or find the original surveyor's notes.

### Thoreau's Principle

A surveyor named Thoreau, working in Massachusetts half a century after my expedition, wrote something that captures this perfectly:

> "It is remarkable how long men will believe in the bottomlessness of a pond without taking the trouble to sound it."

He surveyed Walden Pond himself—over a hundred soundings through the winter ice—because he would not trust hearsay. The sources package embodies this principle: fetch the authoritative data directly.

### The URI System

Just as I learned to distinguish reliable sources from rumor, Chimborazo distinguishes data by URI scheme:

| Scheme | Source | Authority |
|--------|--------|-----------|
| `census://` | U.S. Census Bureau | Federal government TIGER/Line files |
| `nhn://` | Natural Resources Canada | National Hydro Network |
| `file://` | Local filesystem | Your own data |
| `http://` | Remote URL | External GeoJSON or shapefile |

Each scheme knows how to resolve its URI to a downloadable resource, how to parse the result, and how to cache it for future use.

### The Cache

I could not carry sixty thousand specimens back to Paris without organization. Everything was catalogued, preserved, referenced. Years later, I could locate any specimen by its collection number.

The `sources` package maintains a similar cache at `~/.cache/chimborazo/`. Once fetched, data is stored by URI path. Subsequent builds skip the download if the cache is fresh. This is not laziness—it is efficiency. The Census Bureau's servers need not be queried for every test run.

---

## III. Transformation

*Quito, Ecuador, Spring 1802*

Between the acquisition of specimens and the publication of understanding lies the hardest work: *transformation*. I sat for months in Quito, comparing measurements, correlating observations, seeking patterns that would only emerge through systematic comparison.

A single temperature reading tells you nothing. But temperature readings at fifty altitudes, correlated with vegetation observations, with barometric pressure, with latitude—suddenly you see the system. You see that plants arrange themselves by temperature bands, not by arbitrary boundaries. You see that climate, not geography, determines what can grow where.

This is what I mean by "unity in diversity." The raw observations are diverse—chaotic, even. The unity emerges only through transformation.

### Geometric Operations

The `geometry` package performs analogous transformations on spatial data:

**Subtract**: When I mapped the shores of Lake Valencia, I had to distinguish land from water. Subtraction removes one geometry from another—cutting water out of land polygons so that the lake appears as negative space.

**Clip**: My maps could not extend infinitely. I clipped observations to the bounds of my current study—the Orinoco basin, the Quito highlands, the Mexican plateau. Clipping removes everything outside a boundary.

**Merge**: Administrative boundaries are often fragmented—census blocks, municipalities, provinces. Merging combines adjacent geometries into unified wholes.

**Dissolve**: Sometimes the fragments share an attribute—a county code, a province name. Dissolving groups features by attribute and merges each group.

**Simplify**: For publication, some detail must be sacrificed. A coastline with ten thousand vertices becomes unmanageable. Simplification reduces vertex count while preserving essential shape.

### The OperationContext

Every operation receives context: the bounds of the output, the other sources available for reference (needed for subtract operations), the tolerance levels for simplification. This context flows through the transformation pipeline, ensuring that each operation knows what it needs to know.

---

## IV. Revelation

*Paris, 1807*

The Naturgemälde was not merely a diagram. It was a *revelation*.

I drew Chimborazo in cross-section, showing the vegetation zones arranged by altitude. At the base: tropical palms. Higher: oaks and shrubs of the temperate zone. Higher still: alpine grasses. At the heights I reached: bare rock and eternal snow.

But I did not stop there. I added columns showing temperature, barometric pressure, humidity, the boiling point of water at each altitude, the blueness of the sky, the animals observed. Everything on a single page. A viewer could see at a glance what no table of numbers could convey: that these phenomena are *connected*. That altitude, temperature, pressure, and life form a single system.

Florence Kelley understood this principle seventy years later when she mapped Chicago's 19th Ward. She did not merely list wages and nationalities in tables. She mapped them—color-coded by ethnicity, shaded by income level. A viewer could see at a glance the clustering of immigrants, the geography of poverty. "By looking at just eight pages," she wrote, "the reader could easily see."

### Maps Are Evidence

The `output` package produces SVG—Scalable Vector Graphics. This format was chosen deliberately:

- **Precision**: Vector paths maintain accuracy at any zoom level
- **Plotter-ready**: Pen plotters and laser cutters consume SVG directly
- **Layered**: SVG groups preserve the layer structure from the recipe
- **Styleable**: Stroke widths, fills, and colors are explicit attributes

But format is not enough. The output must *reveal*. Every style choice serves comprehension:

| Choice | Purpose |
|--------|---------|
| Stroke width | Hierarchy: state boundaries thicker than county |
| Color | Distinction: water blue, land outline black |
| Layer order | Visibility: roads above everything, background below |
| Simplification | Legibility: too much detail obscures pattern |

The plain citizen who uses these maps deserves data in a form ready to understand.

---

## V. Return

*Paris, 1805-1834*

I spent twenty-three years in Paris after my return, publishing the results of my expedition. Thirty volumes. Hundreds of maps and diagrams. Thousands of pages. I spent my entire fortune on these publications—what would be millions in today's currency.

Why Paris? Because only there could I find the engravers, the printers, the scientific community, the libraries I needed. No single person, no single resource could have produced what I produced. It required *coordination*—the aggregation of many capabilities toward a single goal.

### Maury's Insight

Matthew Fontaine Maury, the American oceanographer, understood this better than anyone. From his office in Washington, he collected ship logs from captains worldwide. No single captain could map the ocean's currents. But thousands of observations, aggregated systematically, revealed patterns invisible to any individual voyage: the Gulf Stream, the trade wind routes, the optimal paths across the Atlantic.

"There is a river in the ocean," he wrote. "In the severest droughts it never fails, and in the mightiest floods it never overflows."

That river only became visible through coordination.

### The Pipeline

The `pipeline` package is Chimborazo's coordinator. It:

1. **Validates** the recipe (Preparation)
2. **Fetches** all required sources, respecting cache (Acquisition)
3. **Processes** each layer through its operations (Transformation)
4. **Renders** the final SVG output (Revelation)

Each stage depends on the previous. You cannot transform what you have not acquired. You cannot render what you have not transformed. The pipeline ensures correct ordering, handles errors gracefully, and reports progress.

### The Build Context

As the pipeline executes, it maintains a `BuildContext`—the accumulated state of the expedition:

```go
type BuildContext struct {
    Recipe          *Recipe
    Cache           *Cache
    SourceData      map[string]FeatureCollection
    ProcessedLayers map[string]FeatureCollection
}
```

Sources are loaded into `SourceData`. Layers are processed into `ProcessedLayers`. The context grows as the build progresses, carrying forward what each stage needs from previous stages.

---

## Epilogue: The Summit We Approach

I never reached the summit of Chimborazo. An impassable crevasse stopped me a thousand feet short. I could see the peak through the clouds, but I could not touch it.

And yet I count that climb among my greatest achievements. From 19,286 feet, I saw what no one had seen before. I understood what no table of numbers could convey. The view was worth the climb, even without the summit.

Every map is like this. We approach perfection asymptotically. The Census Bureau's data has errors. The simplification algorithm loses detail. The SVG renderer makes approximations. The perfect map remains beyond the crevasse.

But the map we produce—the view from where we actually stand—reveals connections that would otherwise stay forever lost. Towns take shape from census blocks. Water carves negative space from land. State boundaries emerge from county fragments. Pattern emerges from data.

This is what Chimborazo offers: not perfection, but revelation. Not the summit, but the view.

---

> *"The most important aim of all physical science is this: to recognize unity in diversity."*
>
> — Alexander von Humboldt, *Cosmos*, 1845

---

## Appendix: The Expedition Timeline

For those who wish to understand the historical journey that informs this metaphor:

| Date | Location | Event |
|------|----------|-------|
| March 1799 | Madrid | Secures royal permission |
| June 1799 | La Coruña | Departs Spain aboard the Pizarro |
| July 1799 | Cumaná, Venezuela | First steps on South American soil |
| Feb-June 1800 | Orinoco River | 1,725-mile journey through wilderness |
| July 1801 | Bogotá, Colombia | Meets botanist José Celestino Mutis |
| June 23, 1802 | Chimborazo, Ecuador | Reaches 19,286 feet—world altitude record |
| 1803-1804 | Mexico | Surveys New Spain |
| May-July 1804 | Philadelphia, Washington D.C. | Meets President Jefferson |
| August 1804 | Paris | Returns to Europe |
| 1805-1834 | Paris | Publishes 30 volumes |
| 1845-1862 | Berlin | Writes *Cosmos* |
| May 6, 1859 | Berlin | Dies at age 89 |

The expedition lasted five years. The publications took thirty. The insights endure.
