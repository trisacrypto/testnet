#!/bin/bash
# TODO: check codes

if [ "$1" == "--local" ]; then
    ALICE_ENDPOINT=localhost:5434
    BOB_ENDPOINT=localhost:6434
else
    ALICE_ENDPOINT=admin.alice.vaspbot.net:443
    BOB_ENDPOINT=admin.bob.vaspbot.net:443
fi

# Send from Alice to Bob
# Partial Sync Repair: success expected
echo "Alice --> Bob Partial Sync Repair"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX -d 0.0001 \
    -b 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh

# Partial Sync Require: rejection expected
echo "Alice --> Bob Partial Sync Require"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX -d 0.0002 \
    -b 1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw

# Full Sync Repair: success expected
echo "Alice --> Bob Full Sync Repair"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha -d 0.0003 \
    -b 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh

# Full Sync Require: success expected
echo "Alice --> Bob Full Sync Require"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v -d 0.0004 \
    -b 1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw

# Partial Async Repair: success expected
# TODO: how to test getting a message back?
echo "Alice --> Bob Partial Async Repair"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX -d 0.0005 \
    -b 14WU745djqecaJ1gmtWQGeMCFim1W5MNp3

# Partial Async Reject: rejection expected
echo "Alice --> Bob Partial Async Reject"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX -d 0.0006 \
    -b 1Hzej6a2VG7C8iCAD5DAdN72cZH5THSMt9

# Full Async Repair: success expected
echo "Alice --> Bob Full Async Repair"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha -d 0.0007 \
    -b 14WU745djqecaJ1gmtWQGeMCFim1W5MNp3

# Full Async Require: reject expected
echo "Alice --> Bob Full Async Reject"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v -d 0.0008 \
    -b 1Hzej6a2VG7C8iCAD5DAdN72cZH5THSMt9

# Send Error: rejection expected.
echo "Alice --> Bob Send Error"
rvasp transfer -e $ALICE_ENDPOINT \
    -a 19nFejdNSUhzkAAdwAvP3wc53o8dL326QQ -d 0.0009 \
    -b 1Hzej6a2VG7C8iCAD5DAdN72cZH5THSMt9

# Send from Bob to Alice
# Partial Sync Repair: success expected
echo "Bob --> Alice Partial Sync Repair"
rvasp transfer -e $BOB_ENDPOINT \
    -a 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh -d 0.00010 \
    -b 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX

# Partial Sync Require: rejection expected
echo "Bob --> Alice Partial Sync Require"
rvasp transfer -e $BOB_ENDPOINT \
    -a 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh -d 0.00011 \
    -b 1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha

# Full Sync Repair: success expected
echo "Bob --> Alice Full Sync Repair"
rvasp transfer -e $BOB_ENDPOINT \
    -a 1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw -d 0.00012 \
    -b 1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX

# Full Sync Require: success expected
echo "Bob --> Alice Full Sync Require"
rvasp transfer -e $BOB_ENDPOINT \
    -a 14WU745djqecaJ1gmtWQGeMCFim1W5MNp3 -d 0.00013 \
    -b 1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha

# Partial Async Repair: success expected
echo "Bob --> Alice Partial Async Repair"
rvasp transfer -e $BOB_ENDPOINT \
    -a 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh -d 0.00014 \
    -b 14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v

# Partial Async Require: rejection expected
echo "Bob --> Alice Partial Async Reject"
rvasp transfer -e $BOB_ENDPOINT \
    -a 18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh -d 0.00015 \
    -b 19nFejdNSUhzkAAdwAvP3wc53o8dL326QQ

# Full Async Repair: success expected
echo "Bob --> Alice Full Async Repair"
rvasp transfer -e $BOB_ENDPOINT \
    -a 1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw -d 0.00016 \
    -b 14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v

# Full Async Require: rejection expected
echo "Bob --> Alice Full Async Reject"
rvasp transfer -e $BOB_ENDPOINT \
    -a 14WU745djqecaJ1gmtWQGeMCFim1W5MNp3 -d 0.00017 \
    -b 19nFejdNSUhzkAAdwAvP3wc53o8dL326QQ

# Send Error: rejection expected.
echo "Bob --> Alice Send Error"
rvasp transfer -e $BOB_ENDPOINT \
    -a 1Hzej6a2VG7C8iCAD5DAdN72cZH5THSMt9 -d 0.00018 \
    -b 14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v