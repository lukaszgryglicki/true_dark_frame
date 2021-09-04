#!/usr/bin/env ruby
require 'rmagick'
def error(s)
  puts s
  exit
end
error 'Please provide file name' unless ARGV.any?
for iname in ARGV do
  imgs = Magick::Image.read(iname)
  error "Cannot read image: #(iname}" unless imgs
  img = imgs[0]
  puts "Read #{iname}"
  sums = [0, 0, 0, 0]
  cnts = [0, 0, 0, 0]
  img.each_pixel do |pixel, c, r|
    i = ((r % 2) << 1) + (c % 2)
    sums[i] += pixel.red
    cnts[i] += 1
  end
  p sums
  p cnts
  avg = sums.inject(0.0) { |sum, el| sum + el } / sums.size
  aval = avg / cnts[0]
  puts "Averaging to: #{avg.to_i} / #{aval.to_i}"
  bias = []
  sums.each { |s| bias << s / avg }
  p bias
  pixels = []
  osums = [0, 0, 0, 0]
  img.each_pixel do |pixel, c, r|
    i = ((r % 2) << 1) + (c % 2)
    v = (pixel.red / bias[i]).to_i
    pixel.red = pixel.green = pixel.blue = v
    osums[i] += pixel.red
    pixels << pixel
  end
  puts "Image recomputed, saving updated pixels"
  img.store_pixels(0, 0, img.columns, img.rows, pixels)
  p osums
  fn = "out_#{iname}"
  img.write(fn)
  puts "Saved: #{fn}"
end
