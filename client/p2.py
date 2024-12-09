
from gameClient import Client
from strategy_examples import FirstMoveStrategy

# URLBASE = "https://rlarena.akbakas.de/"
# client = Client("radomMove", FirstMoveStrategy(),urlbase=URLBASE)


client = Client("firstMove", FirstMoveStrategy())

# After sign up the user will receive a token.
# This token will be used to authenticate the user / client.
# This token will be saved in the .usertoken file.
# If you lose the token by e.g. deleting the file, you can not sign in
# with your old account
# Therefore, do not delete the .usertoken file.
# If you want to create another account, just change the username :)
# Both tokens will be stored! 

success = client.signUp()
if not success:
    print("Sign up failed")
    exit()


ag = client.getActiveGames()

print(client.username, "my trun :{} awating:{}".format(
    len(ag["my_turn"]), len(ag["awaiting"]))
)

client.play(120)
