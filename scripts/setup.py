import setuptools

def requirements(filename):
    with open(filename) as f:
        lines = f.read().splitlines()
    return lines

setuptools.setup(
    name='vrouter-ctl',
    version='0.1',
    packages=setuptools.find_packages(),

    # metadata
    author="OpenContrail",
    author_email="dev@lists.opencontrail.org",
    license="Apache Software License",
    url="http://www.opencontrail.org/",
    long_description="OpenContrail vrouter command-line interface",
    install_requires=requirements('requirements.txt'),

    #test_suite='vrouter_ctl.tests',
    #tests_require=requirements('test-requirements.txt'),

    entry_points = {
        'console_scripts': [
            'vrouter-ctl = vrouter_ctl.vrouter_ctl:main',
        ],
    }
)
