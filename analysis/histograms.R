l5 <- read.csv("histdata/LT05_L1TP_188013_19850709_20200918_02_T1_hist.csv")
l7 <- read.csv("histdata/LE07_L1TP_186013_20000728_20201008_02_T1_hist.csv")
l8 <- read.csv("histdata/LC08_L1TP_187013_20210705_20210713_02_T1_hist.csv")

colnames(l5) <- c("band", "pos", "x", "f")
colnames(l7) <- c("band", "pos", "x", "f")
colnames(l8) <- c("band", "pos", "x", "f")

xx <- rbind(cbind(sat = "L5", l5), cbind(sat = "L7", l7), cbind(sat = "L8", l8))

lapply(split(xx, xx$band), function(x) {
  quartz() # dev.new()
  band <- unique(x$band)
  xlim <- range(x$x)
  for (sat in c("L5", "L7", "L8")) {
    data <- x[x$sat == sat, , drop = FALSE]
    data$f <- data$f / sum(data$f)
    if (sat == "L5") {
      plot(data$x, data$f, type = "l", xlim = xlim, ylim = c(0, 0.2), col = "blue", main = paste("band", band))
    } else if (sat == "L7") {
      lines(data$x, data$f, col = "green")
    } else {
      lines(data$x, data$f, col = "red")
    }
  }
})


# B1: {0.085, 0.85}
