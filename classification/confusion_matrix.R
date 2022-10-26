library(RColorBrewer)

cm <- rbind(
    c(2977, 3, 0, 0, 0, 5, 0, 2, 6, 0, 7),
    c(0, 2409, 2, 10, 0, 112, 27, 54, 175, 208, 3),
    c(0, 152, 789, 278, 2, 61, 39, 76, 46, 419, 81),
    c(0, 42, 21, 1957, 15, 219, 18, 102, 281, 217, 128),
    c(0, 93, 1, 102, 308, 42, 71, 69, 97, 124, 0),
    c(0, 73, 1, 72, 0, 1432, 3, 132, 894, 282, 111),
    c(0, 195, 4, 32, 3, 19, 338, 142, 45, 280, 0),
    c(0, 166, 1, 25, 1, 121, 148, 1281, 860, 386, 11),
    c(0, 56, 0, 79, 1, 454, 15, 361, 1572, 360, 102),
    c(0, 51, 1, 71, 2, 164, 44, 197, 409, 1944, 117),
    c(0, 25, 2, 179, 1, 357, 4, 51, 178, 354, 1849)
)

nms <- c(
    "Cloud",
    "Clean water",
    "Water with sediments",
    "Non-vegetated / Stone tundra",
    "Burnt areas, mostly new",
    "Dwarf shrub",
    "Wetland",
    "Coniferous (pine, damaged spruce)",
    "Coniferous (spruce)",
    "Deciduous",
    "Tundra vegetation"
   )

cl <- colorRampPalette(brewer.pal(8, "Blues"))(25)

pdf("/Volumes/Caffeine/Data/Landsat/results/v11/confusion-matrix.pdf", height=7, width=7)
heatmap(t(cm), Rowv=NA, Colv=NA, scale="col", revC=TRUE, labRow=nms, labCol=nms, margins=c(18,18), col=cl)
dev.off()
