
##########################################################################
## Module Info
##########################################################################

# Version info should be synced from the version.go file
__version_info__ = {
    'major': 1,
    'minor': 1,
    'micro': 3,
    'releaselevel': 'final',
    'serial': 1,
}


##########################################################################
## Helper Functions
##########################################################################

def get_version(short=False):
    """
    Prints the version.
    """
    assert __version_info__['releaselevel'] in ('alpha', 'beta', 'final')
    vers = ["%(major)i.%(minor)i" % __version_info__, ]
    if __version_info__['micro']:
        vers.append(".%(micro)i" % __version_info__)
    if __version_info__['releaselevel'] != 'final' and not short:
        vers.append('%s%i' % (__version_info__['releaselevel'][0],
                              __version_info__['serial']))
    return ''.join(vers)
