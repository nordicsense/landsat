nms <- c("Cloud",
    "Clean water",
    "Water with sediments",
    "Non-vegetated / Stone tundra",
    "Deciduous w/ birch, willow, grass",
    "Burnt areas, mostly new",
    "Deciduous, mostly recovered w/ birch",
    "Coniferous: 40-60% damaged or pine",
    "Coniferous, mostly spruce",
    "Wetland, mostly open",
    "Dwarf shrub, shrub, lichen")

data <- read.csv("/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trainingdata/trainingdata.csv")
data <- data[order(data$clazzid),]
data <- split(data, data$clazzid)
names(data) <- nms

tdata <- read.csv("/Volumes/Caffeine/Data/Landsat/results/v6-11c-7v/trainingdata/trainingdata-test.csv")
tdata <- data[order(tdata$clazzid),]
tdata <- split(tdata, tdata$clazzid)
names(tdata) <- nms

cbind(training=sapply(data, nrow), test=sapply(tdata, nrow))



pdf(file="~/Data/Landsat/analysis/model/trainingdata.pdf", height=8.4, width=11.6)
datasets <- c("band1", "band2", "band3", "band4", "band5", "band7", "ndvi", "nbr", "ndwi") # "nbr2"
par(mar=c(4,10,2,2), mfrow=c(3,3))
for (dsname in datasets) {
    boxplot(lapply(data, "[[", dsname), horizontal=TRUE, las=2, main=dsname)
}
dev.off()
