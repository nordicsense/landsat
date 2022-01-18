l5 <- read.csv("histdata/LT05_L1TP_188013_19850709_20200918_02_T1_histcorr_hist.csv")
l7 <- read.csv("histdata/LE07_L1TP_186013_20000728_20201008_02_T1_histcorr_hist.csv")
l8 <- read.csv("histdata/LC08_L1TP_187013_20210705_20210713_02_T1_histcorr_hist.csv")
l82 <- read.csv("histdata/LC08_L1TP_186013_20130724_20200912_02_T1_histcorr_hist.csv")
l83 <- read.csv("histdata/LC08_L1TP_187012_20170710_20200903_02_T1_histcorr_hist.csv")

colnames(l5) <- c("band", "pos", "x", "f")
colnames(l7) <- c("band", "pos", "x", "f")
colnames(l8) <- c("band", "pos", "x", "f")
colnames(l82) <- c("band", "pos", "x", "f")
colnames(l83) <- c("band", "pos", "x", "f")

xx <- rbind(cbind(sat = "L5", l5), cbind(sat = "L7", l7), cbind(sat = "L8", l8), cbind(sat = "L82", l82), cbind(sat = "L83", l83))

lapply(split(xx, xx$band), function(x) {
  quartz() # dev.new()
  band <- unique(x$band)
  xlim <- range(x$x)
  for (sat in c("L5", "L7", "L8", "L82", "L83")) {
    data <- x[x$sat == sat, , drop = FALSE]
    data$f <- data$f / sum(data$f)
    if (sat == "L5") {
      plot(data$x, data$f, type = "l", xlim = xlim, ylim = c(0, 0.2), col = "blue", main = paste("band", band))
    } else if (sat == "L7") {
      lines(data$x, data$f, col = "green")
    } else if (sat == "L8") {
      lines(data$x, data$f, col = "red")
    } else if (sat == "L82") {
      lines(data$x, data$f, col = "orange")
    } else if (sat == "L83") {
      lines(data$x, data$f, col = "pink")
    }
  }
})
