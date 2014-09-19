# Be sure to restart your server when you modify this file.

# Your secret key is used for verifying the integrity of signed cookies.
# If you change this key, all old signed cookies will become invalid!

# Make sure the secret is at least 30 characters and all random,
# no regular words or you'll be exposed to dictionary attacks.
# You can use `rake secret` to generate a secure secret key.

# Make sure your secret_key_base is kept private
# if you're sharing your code publicly.
CurateNd::Application.config.secret_token = 'a65a9170fbaaa136de99efc80fbb2791b15d3780645534637448a59480c87a50d491aa3c2ad1fdedada5bd4114ff409e84a17a0d9fda6aa89dda2fa1f09aa37f'
CurateNd::Application.config.secret_key_base = 'bab4f5cc222b247d59e19b5c426d9b124c3432ac92c31cf01904f5e4380927cc4b2eddabdc39872a036f5a2b2e51f67854fda7bc131f73c28b4600aa795f126b'