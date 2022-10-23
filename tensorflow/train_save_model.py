import os
import pandas as pd
import tensorflow as tf
import tensorflow.keras.layers as layers
import tensorflow.keras.losses as losses

nClasses = 16

root = os.environ.get("RESULTS_DIR")

df = pd.read_csv(root + '/trainingdata/trainingdata.csv')
x = df.drop(['clazz', 'clazzid'], axis=1)
y = df['clazzid']

df = pd.read_csv(root + '/trainingdata/trainingdata-test.csv')
x_test = df.drop(['clazz', 'clazzid'], axis=1)
y_test = df['clazzid']

tf.random.set_seed(42)

model = tf.keras.Sequential([
    layers.Dense(512, activation=tf.nn.relu, name="outer"),
    layers.Dense(256, activation=tf.nn.relu, name="inner"),
    layers.Dense(nClasses, activation=tf.nn.softmax, name="clazzifier"),
], name="sequential")

model.compile(
    optimizer=tf.keras.optimizers.Adam(learning_rate=0.001),
    loss=losses.SparseCategoricalCrossentropy(reduction=losses.Reduction.SUM),
    metrics=['accuracy']
)

model.fit(x, y, epochs=10)
model.evaluate(x_test, y_test, verbose=2)


# saved_model_cli show --dir tf.model --all
model.save(root + '/tf.model')

